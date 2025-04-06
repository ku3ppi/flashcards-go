# flashcards-go
CLI flashcard app written in go using cobra &amp;&amp; pterm + web UI

A colorful terminal-based flashcard application for effective learning and memorization. Built with Go and pterm for a beautiful terminal UI.

## Features

- Create and manage multiple flashcard decks
- Add cards with questions and answers
- Study mode with spaced repetition
- **Quiz mode with multiple choice questions**
- Beautiful, colorful terminal interface

## Installation

```bash
# Clone the repository
git clone https://github.com/ku3ppi/flashcard-app.git
cd flashcard-app

# Build the application
go build -o flashcard
```

## Usage

```bash
# Create a new deck
./flashcard create --name "Spanish Vocabulary"

# Add cards to a deck
./flashcard add --deck "Spanish Vocabulary"

# List all available decks
./flashcard list

# Study a deck
./flashcard study --deck "Spanish Vocabulary"

# Quiz yourself on a deck (with multiple choice questions)
./flashcard quiz --deck "Spanish Vocabulary" --questions 10
```

## Study Modes

### Standard Study
Review your flashcards one by one, revealing the answer after you've thought about it. You decide if you got it right or wrong, which affects the spaced repetition system.

### Quiz Mode
Test your knowledge with multiple-choice questions! The app will:
- Present questions in random order
- Provide 4 options for each question (1 correct answer + 3 distractors)
- Track your score and display results at the end
- Update card proficiency levels based on your answers

## Spaced Repetition System

This flashcard app uses a simple spaced repetition system with 5 boxes:

- Box 1: Cards you need to review most frequently
- Box 2-4: Intermediate boxes
- Box 5: Cards you know very well

When you answer a card correctly, it moves up one box. If you answer incorrectly, the card goes back to box 1.

## Web Version (Optional)

A web frontend using React, SASS/SCSS, and unoCSS is planned for a future release.