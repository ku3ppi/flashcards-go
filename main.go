package main

import (
	"encoding/json"
	"errors"
	"flag"
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

type Flashcard struct {
	ID             int        `json:"id"`
	Question       string     `json:"question"`
	Answer         string     `json:"answer"`
	CorrectAnswers []string   `json:"correct_answers"`
	Options        []string   `json:"options,omitempty"`
	Category       string     `json:"category"`
	CreatedAt      time.Time  `json:"created_at"`
	LastReviewed   *time.Time `json:"last_reviewed,omitempty"`
	TimesReviewed  int        `json:"times_reviewed"`
	TimesCorrect   int        `json:"times_correct"`
}

type FlashcardApp struct {
	FilePath   string
	Flashcards []Flashcard
	maxID      int
}

func NewFlashcardApp(filePath string) *FlashcardApp {
	app := &FlashcardApp{
		FilePath:   filePath,
		Flashcards: []Flashcard{},
		maxID:      0,
	}
	app.loadFlashcards()
	return app
}

func (app *FlashcardApp) loadFlashcards() error {
	if _, err := os.Stat(app.FilePath); errors.Is(err, os.ErrNotExist) {
		pterm.Warning.Printf("Flashcard file '%s' not found. Starting with an empty set.\n", app.FilePath)
		app.Flashcards = []Flashcard{}
		app.maxID = 0
		return nil
	}

	data, err := ioutil.ReadFile(app.FilePath)
	if err != nil {
		pterm.Error.Printf("Error reading flashcard file '%s': %v\n", app.FilePath, err)
		return err
	}

	if len(data) == 0 {
		pterm.Warning.Printf("Flashcard file '%s' is empty. Starting with an empty set.\n", app.FilePath)
		app.Flashcards = []Flashcard{}
		app.maxID = 0
		return nil
	}

	err = json.Unmarshal(data, &app.Flashcards)
	if err != nil {
		pterm.Error.Printf("Error decoding flashcard JSON from '%s': %v\n", app.FilePath, err)
		pterm.Warning.Println("Could not load existing cards. Starting with an empty set.")
		app.Flashcards = []Flashcard{}
		app.maxID = 0
		return err
	}

	app.maxID = 0
	for _, card := range app.Flashcards {
		if card.ID > app.maxID {
			app.maxID = card.ID
		}
	}
	pterm.Info.Printf("Loaded %d flashcards from '%s'.\n", len(app.Flashcards), app.FilePath)
	return nil
}

func (app *FlashcardApp) saveFlashcards() error {
	data, err := json.MarshalIndent(app.Flashcards, "", "  ")
	if err != nil {
		pterm.Error.Printf("Error encoding flashcards to JSON: %v\n", err)
		return err
	}

	err = ioutil.WriteFile(app.FilePath, data, 0644)
	if err != nil {
		pterm.Error.Printf("Error writing flashcard file '%s': %v\n", app.FilePath, err)
		return err
	}
	return nil
}

func (app *FlashcardApp) getNextID() int {
	app.maxID++
	return app.maxID
}

func (app *FlashcardApp) addCard(question, answer, category string, options, correctAnswers []string) {
	if category == "" {
		category = "General"
	}

	if len(options) > 0 && len(correctAnswers) == 0 {
		if len(options) > 0 {
			correctAnswers = []string{options[0]}
			pterm.Warning.Printf("No correct answer specified for multiple choice. Defaulting to first option: '%s'\n", options[0])
		} else {
			correctAnswers = []string{answer}
		}
	} else if len(options) == 0 {

		correctAnswers = []string{answer}
	}

	newCard := Flashcard{
		ID:             app.getNextID(),
		Question:       question,
		Answer:         answer,
		CorrectAnswers: correctAnswers,
		Options:        options,
		Category:       category,
		CreatedAt:      time.Now(),
		LastReviewed:   nil,
		TimesReviewed:  0,
		TimesCorrect:   0,
	}

	app.Flashcards = append(app.Flashcards, newCard)
	err := app.saveFlashcards()
	if err == nil {
		pterm.Success.Printf("Added new card (ID: %d) to '%s': %s\n", newCard.ID, app.FilePath, question)
	}
}

func (app *FlashcardApp) findCardIndexByID(id int) (int, bool) {
	for i, card := range app.Flashcards {
		if card.ID == id {
			return i, true
		}
	}
	return -1, false
}

func (app *FlashcardApp) reviewCards(categoryFilter string) {
	reviewCards := []Flashcard{}
	if categoryFilter != "" {
		for _, card := range app.Flashcards {
			if strings.EqualFold(card.Category, categoryFilter) {
				reviewCards = append(reviewCards, card)
			}
		}
		pterm.Info.Printf("Reviewing %d cards in category '%s' from '%s'.\n", len(reviewCards), categoryFilter, app.FilePath)
	} else {
		reviewCards = app.Flashcards
		pterm.Info.Printf("Reviewing all %d cards from '%s'.\n", len(reviewCards), app.FilePath)
	}

	if len(reviewCards) == 0 {
		pterm.Warning.Println("No cards to review in this selection.")
		return
	}

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

		if len(card.CorrectAnswers) > 1 {
			pterm.FgLightGreen.Println("\nCorrect answers:")
			for _, ans := range card.CorrectAnswers {
				pterm.FgGreen.Println("- ", ans)
			}
		} else if len(card.CorrectAnswers) == 1 {
			pterm.FgLightGreen.Println("\nAnswer:", card.CorrectAnswers[0])
		} else {
			pterm.FgLightGreen.Println("\nAnswer:", card.Answer)
		}

		result, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultValue(true).
			WithConfirmText("y").
			WithRejectText("n").
			Show("Did you get it right?")

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
		fmt.Println()
	}

	err := app.saveFlashcards()
	if err != nil {
		pterm.Error.Println("Failed to save review results.")
	}

	score := 0.0
	if totalCount > 0 {
		score = (float64(correctCount) / float64(totalCount)) * 100
	}
	pterm.Info.Printf("Review complete! You got %d/%d correct (%.1f%%).\n", correctCount, totalCount, score)
}

func (app *FlashcardApp) quizMode(categoryFilter string, numQuestions int) {
	quizCardsSource := []Flashcard{}
	if categoryFilter != "" {
		for _, card := range app.Flashcards {
			if strings.EqualFold(card.Category, categoryFilter) {
				quizCardsSource = append(quizCardsSource, card)
			}
		}
		pterm.Info.Printf("Starting quiz with cards from category '%s' in '%s'.\n", categoryFilter, app.FilePath)
	} else {
		quizCardsSource = app.Flashcards
		pterm.Info.Printf("Starting quiz with cards from all categories in '%s'.\n", app.FilePath)
	}

	if len(quizCardsSource) == 0 {
		pterm.Warning.Println("No cards available for the quiz in this selection.")
		return
	}
	if numQuestions > len(quizCardsSource) {
		numQuestions = len(quizCardsSource)
		pterm.Info.Printf("Reduced quiz size to %d questions (maximum available).\n", numQuestions)
	}
	if numQuestions <= 0 {
		pterm.Warning.Println("Number of questions must be positive.")
		return
	}

	rand.Shuffle(len(quizCardsSource), func(i, j int) {
		quizCardsSource[i], quizCardsSource[j] = quizCardsSource[j], quizCardsSource[i]
	})
	quizCards := quizCardsSource[:numQuestions]

	correctCount := 0

	pterm.DefaultHeader.Printf("QUIZ MODE: %d questions from %s", numQuestions, app.FilePath)

	for i, card := range quizCards {
		pterm.DefaultSection.Printf("Question %d/%d", i+1, numQuestions)
		pterm.FgLightBlue.Println(card.Question)

		isMultipleChoice := len(card.Options) > 0
		isCorrect := false
		var userAnswer string

		if isMultipleChoice {
			displayOptions := make([]string, len(card.Options))
			copy(displayOptions, card.Options)
			rand.Shuffle(len(displayOptions), func(k, l int) {
				displayOptions[k], displayOptions[l] = displayOptions[l], displayOptions[k]
			})

			optionChoices := []string{}
			for j, option := range displayOptions {
				optionChoices = append(optionChoices, fmt.Sprintf("%d. %s", j+1, option))
			}

			selectedOptionStr, _ := pterm.DefaultInteractiveSelect.
				WithOptions(optionChoices).
				WithDefaultText("Select your answer").
				Show()

			parts := strings.SplitN(selectedOptionStr, ". ", 2)
			if len(parts) == 2 {
				userAnswer = parts[1]
			} else {
				userAnswer = selectedOptionStr
			}

			for _, correctAnswer := range card.CorrectAnswers {
				if strings.EqualFold(userAnswer, correctAnswer) {
					isCorrect = true
					break
				}
			}

		} else {
			userAnswer, _ = pterm.DefaultInteractiveTextInput.Show("Your answer")
			userAnswer = strings.TrimSpace(userAnswer)

			for _, correctAnswer := range card.CorrectAnswers {
				if strings.EqualFold(userAnswer, correctAnswer) {
					isCorrect = true
					break
				}
			}
		}

		originalIndex, found := app.findCardIndexByID(card.ID)
		if found {
			now := time.Now()
			app.Flashcards[originalIndex].TimesReviewed++
			app.Flashcards[originalIndex].LastReviewed = &now
		}

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
				pterm.FgRed.Printf("The correct answer was: %s\n", card.Answer)
			}
		}
		time.Sleep(500 * time.Millisecond)
		fmt.Println()
	}

	err := app.saveFlashcards()
	if err != nil {
		pterm.Error.Println("Failed to save quiz results.")
	}

	score := 0.0
	if numQuestions > 0 {
		score = (float64(correctCount) / float64(numQuestions)) * 100
	}
	pterm.Info.Printf("Quiz complete! You scored %d/%d (%.1f%%).\n", correctCount, numQuestions, score)
}

func (app *FlashcardApp) listCards(categoryFilter string) {
	displayCards := []Flashcard{}
	if categoryFilter != "" {
		for _, card := range app.Flashcards {
			if strings.EqualFold(card.Category, categoryFilter) {
				displayCards = append(displayCards, card)
			}
		}
	} else {
		displayCards = app.Flashcards
	}

	if len(displayCards) == 0 {
		if categoryFilter != "" {
			pterm.Warning.Printf("No cards found in category '%s' in '%s'.\n", categoryFilter, app.FilePath)
		} else {
			pterm.Warning.Printf("No flashcards available in '%s'.\n", app.FilePath)
		}
		return
	}

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

		answerText := card.Answer
		if len(card.CorrectAnswers) > 1 {
			answerText = fmt.Sprintf("%s (+%d more)", card.CorrectAnswers[0], len(card.CorrectAnswers)-1)
		} else if len(card.CorrectAnswers) == 1 {
			answerText = card.CorrectAnswers[0]
		}

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

func (app *FlashcardApp) getCategories() []string {
	categoryMap := make(map[string]bool)
	for _, card := range app.Flashcards {
		categoryMap[card.Category] = true
	}

	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	return categories
}

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
		app.Flashcards = append(app.Flashcards[:indexToDelete], app.Flashcards[indexToDelete+1:]...)
		err := app.saveFlashcards()
		if err == nil {
			pterm.Success.Printf("Deleted card (ID: %d) from '%s': %s\n", cardID, app.FilePath, deletedQuestion)
			return true
		}
	} else {
		pterm.Error.Printf("Card with ID %d not found in '%s'.\n", cardID, app.FilePath)
	}
	return false
}

func (app *FlashcardApp) selectCategory(prompt string, allowAll bool) string {
	categories := app.getCategories()
	if len(categories) == 0 && !allowAll {
		pterm.Warning.Println("No categories available yet.")
		return ""
	}

	options := []string{}
	if allowAll {
		options = append(options, "[All Categories]")
	}
	options = append(options, categories...)

	if len(options) == 0 {
		pterm.Warning.Println("No categories defined yet.")
		return ""
	}

	selected, _ := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultText(prompt).
		Show()

	if allowAll && selected == "[All Categories]" {
		return ""
	}
	if selected == "" {
		pterm.Warning.Println("No category selected.")
		return ""
	}
	return selected
}

func main() {
	filePath := flag.String("file", "flashcards.json", "Path to the flashcards JSON file")

	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	app := NewFlashcardApp(*filePath)

	for {
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
			WithMaxHeight(10).
			WithDefaultText("Select an action").
			Show()

		if selectedOption == "" {
			pterm.Warning.Println("No action selected.")
			continue
		}

		choice := strings.Split(selectedOption, ".")[0]

		switch choice {
		case "1":
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

		case "2":
			if len(app.Flashcards) == 0 {
				pterm.Warning.Println("No cards to review yet. Add some first!")
				continue
			}
			category := app.selectCategory("Select category to review", true)
			app.reviewCards(category)

		case "3":
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

		case "4":
			if len(app.Flashcards) == 0 {
				pterm.Warning.Println("No cards to list yet.")
				continue
			}
			category := app.selectCategory("Select category to list", true)
			app.listCards(category)

		case "5":
			if len(app.Flashcards) == 0 {
				pterm.Warning.Println("No cards to delete.")
				continue
			}
			pterm.Info.Println("Current cards:")
			app.listCards("")

			idStr, _ := pterm.DefaultInteractiveTextInput.
				Show("Enter ID of card to delete")
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err != nil {
				pterm.Error.Println("Invalid ID entered.")
			} else {
				app.deleteCard(id)
			}

		case "6":
			pterm.Info.Println("Goodbye!")
			return

		default:
			pterm.Warning.Println("Invalid selection.")
		}

		fmt.Println()
	}
}
