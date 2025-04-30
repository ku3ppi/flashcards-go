# flashcards-go

A colorful terminal-based flashcard application for effective learning and memorization. Built with Go and [pterm](https://github.com/pterm/pterm) for an interactive terminal UI.

Uses Go's standard `flag` package for command-line argument parsing (like specifying the data file).

## Features

-> Load and save flashcards from/to a JSON file using the `--file` flag. <br>
-> Add cards with questions, answers, categories, and optional multiple-choice options. <br>
-> **Review Mode:** Go through cards one by one and self-assess if you got them right. <br>
-> **Quiz Mode:** Test your knowledge with interactive questions (text input or multiple choice). <br>
-> List existing flashcards, optionally filtered by category. <br>
-> Delete flashcards by ID. <br>
-> Tracks basic statistics (times reviewed, times correct). <br>
-> Interactive terminal interface using pterm. <br>

## Installation

```bash
# Clone the repository (adjust URL if needed)
git clone https://github.com/ku3ppi/flashcards-go.git
cd flashcards-go

# Build the application (optional, you can use go run)
go build -o flashcard-app main.go
```

## Usage

You can run the application directly using `go run` or build it first. The `--file` flag is required to specify the JSON file for storing flashcards.

**Using `go run`:**

```bash
# Run the app, specifying the data file (creates if not found)
# Defaults to 'flashcards.json' in the current directory if --file is omitted
go run main.go --file my_flashcards.json

# Example using an absolute path
go run main.go --file /home/user/Documents/my_cards.json

# Example using the default filename 'flashcards.json'
go run main.go
```

**Using the built executable:**

```bash
# Build it first (if you haven't already)
go build -o flashcard-app main.go

# Run the app, specifying the data file
./flashcard-app --file my_flashcards.json

# Example using an absolute path
./flashcard-app --file /home/user/Documents/my_cards.json

# Example using the default filename 'flashcards.json'
./flashcard-app
```

Once the application starts, follow the interactive menu prompts:

1.  **Add new flashcard:** Enter question, answer, category, and optionally define multiple-choice options.
2.  **Review flashcards:** Go through cards (all or by category) and mark if you answered correctly.
3.  **Quiz mode:** Answer a set number of questions (all or by category) interactively.
4.  **List flashcards:** View a table of your cards (all or by category).
5.  **Delete a flashcard:** Remove a card using its ID after listing them.
6.  **Exit:** Save changes (if any) to the JSON file and close the application.

## Data Storage

Flashcards are stored in a simple JSON file specified by the `--file` flag (or `flashcards.json` by default). If the file doesn't exist at the specified path when the application starts, it will begin with an empty set and create the file upon saving (e.g., after adding a card or finishing a review/quiz).
