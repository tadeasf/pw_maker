package utils

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zalando/go-keyring"
)

var (
	IncludeSpecial bool
	Length         int
	DB             *sql.DB
	DBPath         string
	EncryptionKey  string
)

func InitDB() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(StyleError.Render("Error getting user home directory: " + err.Error()))
		os.Exit(1)
	}

	DBPath = filepath.Join(homeDir, ".fortpass", "passwords.db")
	err = os.MkdirAll(filepath.Dir(DBPath), 0700)
	if err != nil {
		fmt.Println(StyleError.Render("Error creating directory: " + err.Error()))
		os.Exit(1)
	}

	// Retrieve encryption key from system keyring
	EncryptionKey, err = keyring.Get("fortpass", "db_encryption_key")
	if err != nil {
		// Generate a new encryption key if not found
		EncryptionKey = GenerateEncryptionKey()
		err = keyring.Set("fortpass", "db_encryption_key", EncryptionKey)
		if err != nil {
			fmt.Println(StyleError.Render("Error storing encryption key: " + err.Error()))
			os.Exit(1)
		}
	}

	// Open the encrypted database
	DB, err = sql.Open("sqlite3", fmt.Sprintf("%s?_pragma_key=%s", DBPath, EncryptionKey))
	if err != nil {
		fmt.Println(StyleError.Render("Error opening database: " + err.Error()))
		os.Exit(1)
	}

	CreateTable()
}

func GenerateEncryptionKey() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := make([]byte, 32)
	_, err := rng.Read(key)
	if err != nil {
		fmt.Println(StyleError.Render("Error generating encryption key: " + err.Error()))
		os.Exit(1)
	}
	return hex.EncodeToString(key)
}

func CreateTable() {
	_, err := DB.Exec(`
        CREATE TABLE IF NOT EXISTS version (
            version INTEGER PRIMARY KEY
        );
        CREATE TABLE IF NOT EXISTS passwords (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            source TEXT,
            username TEXT,
            password TEXT,
            url TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(source, username, url)
        )
    `)
	if err != nil {
		fmt.Println(StyleError.Render("Error creating tables: " + err.Error()))
		os.Exit(1)
	}

	// Check and update database version
	CheckAndMigrateDatabase()
}

func CheckAndMigrateDatabase() {
	var version int
	err := DB.QueryRow("SELECT version FROM version").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			// Initialize version
			_, err = DB.Exec("INSERT INTO version (version) VALUES (1)")
			if err != nil {
				fmt.Println(StyleError.Render("Error initializing database version: " + err.Error()))
				os.Exit(1)
			}
			version = 1
		} else {
			fmt.Println(StyleError.Render("Error checking database version: " + err.Error()))
			os.Exit(1)
		}
	}

	if version < 2 {
		// Perform migration to version 2
		_, err = DB.Exec(`
            ALTER TABLE passwords ADD COLUMN created_at DATETIME DEFAULT CURRENT_TIMESTAMP;
            ALTER TABLE passwords ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;
        `)
		if err != nil {
			fmt.Println(StyleError.Render("Error migrating database to version 2: " + err.Error()))
			os.Exit(1)
		}
		_, err = DB.Exec("UPDATE version SET version = 2")
		if err != nil {
			fmt.Println(StyleError.Render("Error updating database version: " + err.Error()))
			os.Exit(1)
		}
		fmt.Println(StyleSuccess.Render("Database migrated to version 2"))
	}
}

func GetPasswordEntries() []PasswordEntry {
	rows, err := DB.Query("SELECT source, username, url, created_at, updated_at FROM passwords")
	if err != nil {
		fmt.Println(StyleError.Render("Error fetching passwords: " + err.Error()))
		return nil
	}
	defer rows.Close()

	var entries []PasswordEntry
	for rows.Next() {
		var entry PasswordEntry
		err := rows.Scan(&entry.Source, &entry.Username, &entry.URL, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			fmt.Println(StyleError.Render("Error scanning row: " + err.Error()))
			continue
		}
		entries = append(entries, entry)
	}

	return entries
}
