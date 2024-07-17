# FortPass

FortPass is a secure and user-friendly command-line password manager built in Go. It allows you to generate, store, retrieve, and manage passwords with ease.

## Features

- Generate strong, random passwords
- Store passwords securely in an encrypted SQLite database
- Search and retrieve passwords
- Import passwords from CSV files
- Update and delete existing passwords
- Copy passwords to clipboard with automatic clearing
- User-friendly interface with colorful output

## Installation

1. Ensure you have Go installed on your system.
2. Clone this repository:

   ```sh
   git clone https://github.com/yourusername/fortpass.git
   ```

3. Navigate to the project directory:

   ```sh
   cd fortpass
   ```

4. Build the project:

   ```sh
   go build
   ```

## Usage

Run `./fortpass` followed by a command. Available commands:

- `generate`: Generate a new password
- `show`: Show all stored passwords
- `search`: Search for stored passwords
- `get [source/username]`: Get a specific password by source/username
- `import [csv_file]`: Import passwords from a CSV file
- `delete [source/username]`: Delete a specific password
- `update [source/username]`: Update a specific password

### Examples

Generate a password:

```sh
./fortpass generate -l 16 -s
```

Store a password:

```sh
./fortpass
```

(Follow the prompts to enter details)

Search passwords:

```sh
./fortpass search
```

Get a specific password:

```sh
./fortpass get github.com/johndoe
```

Import passwords from CSV:

```sh
./fortpass import passwords.csv
```

## Security

- Passwords are stored in an encrypted SQLite database
- The database encryption key is securely stored in the system keyring
- Passwords copied to clipboard are automatically cleared after 45 seconds

## Dependencies

- github.com/atotto/clipboard
- github.com/charmbracelet/bubbles
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/lipgloss
- github.com/mattn/go-sqlite3
- github.com/spf13/cobra
- github.com/zalando/go-keyring

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
