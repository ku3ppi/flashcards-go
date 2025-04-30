package main

import (
	"encoding/json"
	"errors"
	"flag" // <-- NEU: Importiere das flag Paket
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

// Flashcard Struct defines the structure for a single flashcard
type Flashcard struct {
	ID             int        `json:"id"`
	Question       string     `json:"question"`
	Answer         string     `json:"answer"` // Main answer text, also used if not multiple choice
	CorrectAnswers []string   `json:"correct_answers"`
	Options        []string   `json:"options,omitempty"` // Options for multiple choice
	Category       string     `json:"category"`
	CreatedAt      time.Time  `json:"created_at"`
	LastReviewed   *time.Time `json:"last_reviewed,omitempty"`
	TimesReviewed  int        `json:"times_reviewed"`
	TimesCorrect   int        `json:"times_correct"`
}

// FlashcardApp holds the application state
type FlashcardApp struct {
	FilePath   string
	Flashcards []Flashcard
	maxID      int // Track the highest ID to avoid reuse after deletion
}

// NewFlashcardApp creates a new instance of the FlashcardApp
func NewFlashcardApp(filePath string) *FlashcardApp {
	app := &FlashcardApp{
		FilePath:   filePath, // <-- GEÄNDERT: Wird jetzt dynamisch übergeben
		Flashcards: []Flashcard{},
		maxID:      0,
	}
	app.loadFlashcards() // Load cards on initialization
	return app
}

// loadFlashcards loads cards from the JSON file
func (app *FlashcardApp) loadFlashcards() error {
	// Check if file exists
	if _, err := os.Stat(app.FilePath); errors.Is(err, os.ErrNotExist) {
		// GEÄNDERT: Zeige den Dateipfad in der Meldung
		pterm.Warning.Printf("Flashcard file '%s' not found. Starting with an empty set.\n", app.FilePath)
		app.Flashcards = []Flashcard{}
		app.maxID = 0
		return nil // Not an error if the file doesn't exist yet
	}

	// Read file
	data, err := ioutil.ReadFile(app.FilePath)
	if err != nil {
		pterm.Error.Printf("Error reading flashcard file '%s': %v\n", app.FilePath, err)
		return err
	}

	// Handle empty file case
	if len(data) == 0 {
		// GEÄNDERT: Zeige den Dateipfad
		pterm.Warning.Printf("Flashcard file '%s' is empty. Starting with an empty set.\n", app.FilePath)
		app.Flashcards = []Flashcard{}
		app.maxID = 0
		return nil
	}

	// Unmarshal JSON
	err = json.Unmarshal(data, &app.Flashcards)
	if err != nil {
		pterm.Error.Printf("Error decoding flashcard JSON from '%s': %v\n", app.FilePath, err)
		pterm.Warning.Println("Could not load existing cards. Starting with an empty set.")
		app.Flashcards = []Flashcard{} // Reset to empty on decode error
		app.maxID = 0
		return err
	}

	// Find the current maximum ID after loading
	app.maxID = 0
	for _, card := range app.Flashcards {
		if card.ID > app.maxID {
			app.maxID = card.ID
		}
	}
	// GEÄNDERT: Zeige den Dateipfad
	pterm.Info.Printf("Loaded %d flashcards from '%s'.\n", len(app.Flashcards), app.FilePath)
	return nil
}

// saveFlashcards saves the current set of cards to the JSON file
func (app *FlashcardApp) saveFlashcards() error {
	data, err := json.MarshalIndent(app.Flashcards, "", "  ") // Pretty print JSON
	if err != nil {
		pterm.Error.Printf("Error encoding flashcards to JSON: %v\n", err)
		return err
	}

	err = ioutil.WriteFile(app.FilePath, data, 0644) // rw-r--r-- permissions
	if err != nil {
		// GEÄNDERT: Zeige den Dateipfad
		pterm.Error.Printf("Error writing flashcard file '%s': %v\n", app.FilePath, err)
		return err
	}
	return nil
}

// getNextID determines the next available ID
func (app *FlashcardApp) getNextID() int {
	app.maxID++
	return app.maxID
}

// addCard adds a new flashcard to the set
func (app *FlashcardApp) addCard(question, answer, category string, options, correctAnswers []string) {
	if category == "" {
		category = "General" // Default category
	}

	// Ensure correctAnswers is populated correctly
	if len(options) > 0 && len(correctAnswers) == 0 {
		// If MC and no correct answers provided, default to first option (or the main answer if available)
		if len(options) > 0 {
			correctAnswers = []string{options[0]}
			pterm.Warning.Printf("No correct answer specified for multiple choice. Defaulting to first option: '%s'\n", options[0])
		} else {
			correctAnswers = []string{answer} // Fallback though unlikely if options exist
		}
	} else if len(options) == 0 {
		// If not MC, the single answer is the correct answer
		correctAnswers = []string{answer}
	}

	newCard := Flashcard{
		ID:             app.getNextID(),
		Question:       question,
		Answer:         answer, // Store main answer text regardless of type
		CorrectAnswers: correctAnswers,
		Options:        options,
		Category:       category,
		CreatedAt:      time.Now(),
		LastReviewed:   nil, // Initialize as nil
		TimesReviewed:  0,
		TimesCorrect:   0,
	}

	app.Flashcards = append(app.Flashcards, newCard)
	err := app.saveFlashcards()
	if err == nil {
		// GEÄNDERT: Zeige Dateipfad
		pterm.Success.Printf("Added new card (ID: %d) to '%s': %s\n", newCard.ID, app.FilePath, question)
	}
}

// findCardIndexByID finds the index of a card in the app.Flashcards slice
func (app *FlashcardApp) findCardIndexByID(id int) (int, bool) {
	for i, card := range app.Flashcards {
		if card.ID == id {
			return i, true
		}
	}
	return -1, false
}

// reviewCards starts a review session
func (app *FlashcardApp) reviewCards(categoryFilter string) {
	reviewCards := []Flashcard{}
	if categoryFilter != "" {
		for _, card := range app.Flashcards {
			if strings.EqualFold(card.Category, categoryFilter) { // Case-insensitive category check
				reviewCards = append(reviewCards, card)
			}
		}
		// GEÄNDERT: Zeige Dateipfad
		pterm.Info.Printf("Reviewing %d cards in category '%s' from '%s'.\n", len(reviewCards), categoryFilter, app.FilePath)
	} else {
		reviewCards = app.Flashcards
		// GEÄNDERT: Zeige Dateipfad
		pterm.Info.Printf("Reviewing all %d cards from '%s'.\n", len(reviewCards), app.FilePath)
	}

	if len(reviewCards) == 0 {
		pterm.Warning.Println("No cards to review in this selection.")
		return
	}

	// Shuffle the review cards
	rand.Shuffle(len(reviewCards), func(i, j int) {
		reviewCards[i], reviewCards[j] = reviewCards[j], reviewCards[i]
	})

	correctCount := 0
	totalCount := len(reviewCards)

	for i, card := range reviewCards {
		pterm.DefaultSection.Printf("Card %d/%d - Category: %s", i+1, totalCount, card.Category)
		pterm.FgLightBlue.Println("Question: ", card.Question)

		isMultipleChoice := len(card.Options) > 0

		if isMultipleChoice {
			pterm.FgYellow.Println("\n(Multiple Choice Question)")
			_, _ = pterm.DefaultInteractiveContinue.Show("Press Enter to see answer options...")

			// Shuffle options for display only
			displayOptions := make([]string, len(card.Options))
			copy(displayOptions, card.Options)
			rand.Shuffle(len(displayOptions), func(k, l int) {
				displayOptions[k], displayOptions[l] = displayOptions[l], displayOptions[k]
			})

			for j, option := range displayOptions {
				pterm.FgCyan.Printf("%d. %s\n", j+1, option)
			}
			_, _ = pterm.DefaultInteractiveContinue.Show("Press Enter to see the correct answer(s)...")
		} else {
			_, _ = pterm.DefaultInteractiveContinue.Show("Press Enter to see the answer...")
		}

		// Display correct answer(s)
		if len(card.CorrectAnswers) > 1 {
			pterm.FgLightGreen.Println("\nCorrect answers:")
			for _, ans := range card.CorrectAnswers {
				pterm.FgGreen.Println("- ", ans)
			}
		} else if len(card.CorrectAnswers) == 1 {
			pterm.FgLightGreen.Println("\nAnswer:", card.CorrectAnswers[0])
		} else {
			// Fallback if CorrectAnswers somehow empty (shouldn't happen with addCard logic)
			pterm.FgLightGreen.Println("\nAnswer:", card.Answer)
		}

		// Ask if correct
		result, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultValue(true).
			WithConfirmText("y").
			WithRejectText("n").
			Show("Did you get it right?") // Prompt directly in Show()

		// Find the original card in the main list to update its stats
		originalIndex, found := app.findCardIndexByID(card.ID)
		if found {
			now := time.Now()
			app.Flashcards[originalIndex].TimesReviewed++
			app.Flashcards[originalIndex].LastReviewed = &now
			if result {
				correctCount++
				app.Flashcards[originalIndex].TimesCorrect++
				pterm.Success.Println("Marked as correct!")
			} else {
				pterm.Warning.Println("Marked as incorrect.")
			}
		} else {
			pterm.Error.Printf("Could not find card with ID %d in main list to update stats.\n", card.ID)
		}
		fmt.Println() // Add a blank line for separation
	}

	// Save changes after the review session
	err := app.saveFlashcards()
	if err != nil {
		pterm.Error.Println("Failed to save review results.")
	}

	// Final score
	score := 0.0
	if totalCount > 0 {
		score = (float64(correctCount) / float64(totalCount)) * 100
	}
	pterm.Info.Printf("Review complete! You got %d/%d correct (%.1f%%).\n", correctCount, totalCount, score)
}

// quizMode starts a quiz session
func (app *FlashcardApp) quizMode(categoryFilter string, numQuestions int) {
	quizCardsSource := []Flashcard{}
	if categoryFilter != "" {
		for _, card := range app.Flashcards {
			if strings.EqualFold(card.Category, categoryFilter) { // Case-insensitive
				quizCardsSource = append(quizCardsSource, card)
			}
		}
		// GEÄNDERT: Zeige Dateipfad
		pterm.Info.Printf("Starting quiz with cards from category '%s' in '%s'.\n", categoryFilter, app.FilePath)
	} else {
		quizCardsSource = app.Flashcards
		// GEÄNDERT: Zeige Dateipfad
		pterm.Info.Printf("Starting quiz with cards from all categories in '%s'.\n", app.FilePath)
	}

	if len(quizCardsSource) == 0 {
		pterm.Warning.Println("No cards available for the quiz in this selection.")
		return
	}

	// Clamp numQuestions to the number of available cards
	if numQuestions > len(quizCardsSource) {
		numQuestions = len(quizCardsSource)
		pterm.Info.Printf("Reduced quiz size to %d questions (maximum available).\n", numQuestions)
	}
	if numQuestions <= 0 {
		pterm.Warning.Println("Number of questions must be positive.")
		return
	}

	// Get a random sample
	rand.Shuffle(len(quizCardsSource), func(i, j int) {
		quizCardsSource[i], quizCardsSource[j] = quizCardsSource[j], quizCardsSource[i]
	})
	quizCards := quizCardsSource[:numQuestions]

	correctCount := 0

	// GEÄNDERT: Zeige Dateipfad im Header
	pterm.DefaultHeader.Printf("QUIZ MODE: %d questions from %s", numQuestions, app.FilePath)

	for i, card := range quizCards {
		pterm.DefaultSection.Printf("Question %d/%d", i+1, numQuestions)
		pterm.FgLightBlue.Println(card.Question)

		isMultipleChoice := len(card.Options) > 0
		isCorrect := false
		var userAnswer string

		if isMultipleChoice {
			// Shuffle options for display
			displayOptions := make([]string, len(card.Options))
			copy(displayOptions, card.Options)
			rand.Shuffle(len(displayOptions), func(k, l int) {
				displayOptions[k], displayOptions[l] = displayOptions[l], displayOptions[k]
			})

			optionChoices := []string{}
			for j, option := range displayOptions {
				optionChoices = append(optionChoices, fmt.Sprintf("%d. %s", j+1, option))
			}

			// Use InteractiveSelect for multiple choice
			selectedOptionStr, _ := pterm.DefaultInteractiveSelect.
				WithOptions(optionChoices).
				WithDefaultText("Select your answer").
				Show()

			// Extract the actual answer text from the selected string (e.g., "1. Option A" -> "Option A")
			parts := strings.SplitN(selectedOptionStr, ". ", 2)
			if len(parts) == 2 {
				userAnswer = parts[1]
			} else {
				userAnswer = selectedOptionStr // Fallback if split fails
			}

			// Check correctness (case-insensitive)
			for _, correctAnswer := range card.CorrectAnswers {
				if strings.EqualFold(userAnswer, correctAnswer) {
					isCorrect = true
					break
				}
			}

		} else {
			// Text input for non-multiple choice
			userAnswer, _ = pterm.DefaultInteractiveTextInput.Show("Your answer")
			userAnswer = strings.TrimSpace(userAnswer)

			// Check correctness (case-insensitive)
			for _, correctAnswer := range card.CorrectAnswers {
				if strings.EqualFold(userAnswer, correctAnswer) {
					isCorrect = true
					break
				}
			}
		}

		// Find the original card to update stats
		originalIndex, found := app.findCardIndexByID(card.ID)
		if found {
			now := time.Now()
			app.Flashcards[originalIndex].TimesReviewed++
			app.Flashcards[originalIndex].LastReviewed = &now
		}

		// Provide feedback
		if isCorrect {
			pterm.Success.Println("Correct! ✓")
			correctCount++
			if found {
				app.Flashcards[originalIndex].TimesCorrect++
			}
		} else {
			pterm.Error.Print("Incorrect. ")
			if len(card.CorrectAnswers) > 1 {
				pterm.FgRed.Printf("The correct answers were: %s\n", strings.Join(card.CorrectAnswers, ", "))
			} else if len(card.CorrectAnswers) == 1 {
				pterm.FgRed.Printf("The correct answer was: %s\n", card.CorrectAnswers[0])
			} else {
				// Fallback
				pterm.FgRed.Printf("The correct answer was: %s\n", card.Answer)
			}
		}
		// Simple pause
		time.Sleep(500 * time.Millisecond)
		fmt.Println()
	}

	// Save changes after the quiz
	err := app.saveFlashcards()
	if err != nil {
		pterm.Error.Println("Failed to save quiz results.")
	}

	// Final score
	score := 0.0
	if numQuestions > 0 {
		score = (float64(correctCount) / float64(numQuestions)) * 100
	}
	pterm.Info.Printf("Quiz complete! You scored %d/%d (%.1f%%).\n", correctCount, numQuestions, score)
}

// listCards displays cards in a table
func (app *FlashcardApp) listCards(categoryFilter string) {
	displayCards := []Flashcard{}
	if categoryFilter != "" {
		for _, card := range app.Flashcards {
			if strings.EqualFold(card.Category, categoryFilter) { // Case-insensitive
				displayCards = append(displayCards, card)
			}
		}
	} else {
		displayCards = app.Flashcards
	}

	if len(displayCards) == 0 {
		if categoryFilter != "" {
			// GEÄNDERT: Zeige Dateipfad
			pterm.Warning.Printf("No cards found in category '%s' in '%s'.\n", categoryFilter, app.FilePath)
		} else {
			// GEÄNDERT: Zeige Dateipfad
			pterm.Warning.Printf("No flashcards available in '%s'.\n", app.FilePath)
		}
		return
	}

	// Sort cards by ID for consistent display
	sort.SliceStable(displayCards, func(i, j int) bool {
		return displayCards[i].ID < displayCards[j].ID
	})

	tableData := pterm.TableData{
		{"ID", "Category", "Question", "Answer(s)", "Type", "Reviewed", "Correct %"},
	}

	for _, card := range displayCards {
		cardType := "Text"
		if len(card.Options) > 0 {
			cardType = "Multiple Choice"
		}

		answerText := card.Answer // Default/first answer
		if len(card.CorrectAnswers) > 1 {
			answerText = fmt.Sprintf("%s (+%d more)", card.CorrectAnswers[0], len(card.CorrectAnswers)-1)
		} else if len(card.CorrectAnswers) == 1 {
			answerText = card.CorrectAnswers[0]
		}

		// Truncate long text for display
		qShort := card.Question
		if len(qShort) > 40 {
			qShort = qShort[:37] + "..."
		}
		aShort := answerText
		if len(aShort) > 30 {
			aShort = aShort[:27] + "..."
		}
		catShort := card.Category
		if len(catShort) > 15 {
			catShort = catShort[:12] + "..."
		}

		reviewedCount := fmt.Sprintf("%d", card.TimesReviewed)
		correctPercent := "N/A"
		if card.TimesReviewed > 0 {
			percent := (float64(card.TimesCorrect) / float64(card.TimesReviewed)) * 100
			correctPercent = fmt.Sprintf("%.0f%%", percent)
		}

		tableData = append(tableData, []string{
			strconv.Itoa(card.ID),
			catShort,
			qShort,
			aShort,
			cardType,
			reviewedCount,
			correctPercent,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}

// getCategories returns a unique list of categories
func (app *FlashcardApp) getCategories() []string {
	categoryMap := make(map[string]bool)
	for _, card := range app.Flashcards {
		categoryMap[card.Category] = true
	}

	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}
	sort.Strings(categories) // Sort for consistent display
	return categories
}

// deleteCard removes a card by its ID
func (app *FlashcardApp) deleteCard(cardID int) bool {
	indexToDelete := -1
	var deletedQuestion string
	for i, card := range app.Flashcards {
		if card.ID == cardID {
			indexToDelete = i
			deletedQuestion = card.Question
			break
		}
	}

	if indexToDelete != -1 {
		// Remove the element from the slice
		app.Flashcards = append(app.Flashcards[:indexToDelete], app.Flashcards[indexToDelete+1:]...)
		err := app.saveFlashcards()
		if err == nil {
			// GEÄNDERT: Zeige Dateipfad
			pterm.Success.Printf("Deleted card (ID: %d) from '%s': %s\n", cardID, app.FilePath, deletedQuestion)
			return true
		}
	} else {
		// GEÄNDERT: Zeige Dateipfad
		pterm.Error.Printf("Card with ID %d not found in '%s'.\n", cardID, app.FilePath)
	}
	return false
}

// selectCategory prompts the user to select a category or choose all
func (app *FlashcardApp) selectCategory(prompt string, allowAll bool) string {
	categories := app.getCategories()
	if len(categories) == 0 && !allowAll { // If no categories and can't select all, warn
		pterm.Warning.Println("No categories available yet.")
		return "" // Return empty string if no categories exist
	}

	options := []string{}
	if allowAll {
		options = append(options, "[All Categories]")
	}
	options = append(options, categories...)

	// Check if only "[All Categories]" is an option when no actual categories exist
	if len(options) == 0 {
		pterm.Warning.Println("No categories defined yet.")
		return "" // Cannot select anything
	}

	selected, _ := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultText(prompt).
		Show()

	if allowAll && selected == "[All Categories]" {
		return "" // Represent "All" with an empty string
	}
	// Handle case where user might have cancelled the select prompt (selected is empty)
	if selected == "" {
		pterm.Warning.Println("No category selected.")
		return "" // Treat cancellation like selecting nothing specific
	}
	return selected // Return the chosen category name
}

// --- Main Application Logic ---
func main() {
	// --- NEU: Flag Definition ---
	// Definiere das Flag: Name "file", Standardwert "flashcards.json", Beschreibung für -h/--help
	filePath := flag.String("file", "flashcards.json", "Path to the flashcards JSON file")

	// Parse die Kommandozeilen-Flags
	flag.Parse()
	// --- Ende NEU ---

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Setup application
	// NEU: Verwende den Wert aus dem Flag (Dereferenzierung mit *)
	app := NewFlashcardApp(*filePath)

	// Main menu loop
	for {
		// NEU/GEÄNDERT: Zeige den verwendeten Dateipfad im Header an
		pterm.DefaultHeader.Printf("=== GO FLASHCARD APP ('%s') ===", app.FilePath)
		options := []string{
			"1. Add new flashcard",
			"2. Review flashcards",
			"3. Quiz mode",
			"4. List flashcards",
			"5. Delete a flashcard",
			"6. Exit",
		}
		selectedOption, _ := pterm.DefaultInteractiveSelect.
			WithOptions(options).
			WithMaxHeight(10). // Deine Ergänzung für die Höhe
			WithDefaultText("Select an action").
			Show()

		// Handle potential cancellation of the main menu selection
		if selectedOption == "" {
			pterm.Warning.Println("No action selected.")
			continue // Go back to the start of the loop
		}

		// Extract choice number (or handle potential errors if format changes)
		choice := strings.Split(selectedOption, ".")[0]

		switch choice {
		case "1": // Add card
			question, _ := pterm.DefaultInteractiveTextInput.Show("Enter question")
			answer, _ := pterm.DefaultInteractiveTextInput.Show("Enter the 'main' answer (used if not multiple choice)")
			category, _ := pterm.DefaultInteractiveTextInput.Show("Enter category (leave blank for 'General')")

			isMultipleChoice, _ := pterm.DefaultInteractiveConfirm.
				WithConfirmText("y").WithRejectText("n").
				Show("Make this a multiple choice question?")

			var mcOptions []string
			var mcCorrectAnswers []string

			if isMultipleChoice {
				pterm.Info.Println("Enter options (type 'done' when finished, need at least 2):")
				optionCount := 1
				for {
					optionText, _ := pterm.DefaultInteractiveTextInput.
						Show(fmt.Sprintf("Option %d", optionCount))

					trimmedOption := strings.ToLower(strings.TrimSpace(optionText))
					if trimmedOption == "done" {
						if len(mcOptions) < 2 {
							pterm.Warning.Println("Need at least 2 options for multiple choice. Please add more.")
							continue
						}
						break
					}

					if optionText != "" {
						mcOptions = append(mcOptions, optionText)
						isCorrect, _ := pterm.DefaultInteractiveConfirm.
							WithConfirmText("y").WithRejectText("n").
							Show(fmt.Sprintf("Is '%s' a correct answer?", optionText))
						if isCorrect {
							mcCorrectAnswers = append(mcCorrectAnswers, optionText)
						}
						optionCount++
					} else if trimmedOption != "done" {
						pterm.Warning.Println("Option cannot be empty. Please enter text or type 'done'.")
					}
				}
			}

			app.addCard(question, answer, category, mcOptions, mcCorrectAnswers)

		case "2": // Review cards
			if len(app.Flashcards) == 0 {
				pterm.Warning.Println("No cards to review yet. Add some first!")
				continue
			}
			category := app.selectCategory("Select category to review", true)
			app.reviewCards(category)

		case "3": // Quiz mode
			if len(app.Flashcards) == 0 {
				pterm.Warning.Println("No cards for a quiz yet. Add some first!")
				continue
			}
			category := app.selectCategory("Select category for quiz", true)

			numStr, _ := pterm.DefaultInteractiveTextInput.
				WithDefaultValue("5").
				Show("Number of questions")

			num, err := strconv.Atoi(strings.TrimSpace(numStr))
			if err != nil || num <= 0 {
				pterm.Warning.Println("Invalid number of questions, defaulting to 5.")
				num = 5
			}
			app.quizMode(category, num)

		case "4": // List cards
			if len(app.Flashcards) == 0 {
				pterm.Warning.Println("No cards to list yet.")
				continue
			}
			category := app.selectCategory("Select category to list", true)
			app.listCards(category)

		case "5": // Delete card
			if len(app.Flashcards) == 0 {
				pterm.Warning.Println("No cards to delete.")
				continue
			}
			pterm.Info.Println("Current cards:")
			app.listCards("") // List all cards so user can see IDs

			idStr, _ := pterm.DefaultInteractiveTextInput.
				Show("Enter ID of card to delete")
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err != nil {
				pterm.Error.Println("Invalid ID entered.")
			} else {
				app.deleteCard(id)
			}

		case "6": // Exit
			pterm.Info.Println("Goodbye!")
			return // Exit the main function

		default:
			pterm.Warning.Println("Invalid selection.")
		}

		fmt.Println() // Add space before next menu iteration
	}
}
