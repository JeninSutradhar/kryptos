package ui

import (
	"fmt"
	"image/color"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"kryptos/entry"
	"kryptos/storage"
)

const (
	paddingSmall       = 5
	paddingMedium      = 10
	paddingLarge       = 20
	accountRowHeight   = 80
	accountCardPadding = 12
)

type UIState struct {
	App             fyne.App
	MainWindow      fyne.Window
	Accounts        []string              // List of account names
	CurrentAccount  string                // Currently selected account
	Entries         []entry.PasswordEntry // Current user's password entries
	MasterPassword  string
	EntryList       *widget.List
	AccountList     *widget.List
	SearchEntry     *widget.Entry
	CurrentScreen   string
	OriginalEntries []entry.PasswordEntry // Store original entries for search filtering
}

var state UIState

func RunUI() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Fatal error in UI: %v", r)
			// You might want to show a dialog here if possible
			panic(r) // Re-panic after logging
		}
	}()

	a := app.New()
	a.Settings().SetTheme(&RoyalDarkTheme{})
	state = UIState{App: a, MainWindow: a.NewWindow("Kryptos Password Manager"), CurrentScreen: "accounts"}
	state.Accounts = storage.LoadAccounts()

	showAccountsScreen()

	state.MainWindow.Resize(fyne.NewSize(1000, 700))
	state.MainWindow.CenterOnScreen()
	state.MainWindow.ShowAndRun()
}
func showAccountsScreen() {
	state.CurrentScreen = "accounts"
	state.MainWindow.SetTitle("Kryptos - Account Manager")

	// Ensure state.AccountList is initialized
	if state.AccountList == nil {
		state.AccountList = widget.NewList(
			func() int { return len(state.Accounts) },
			func() fyne.CanvasObject {
				// Create icon
				icon := widget.NewIcon(theme.AccountIcon())

				// Create labels
				accountLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				subtitleLabel := widget.NewLabelWithStyle("Click to login", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

				// Create buttons
				loginBtn := widget.NewButtonWithIcon("Login", theme.LoginIcon(), nil)
				loginBtn.Importance = widget.HighImportance

				deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
				deleteBtn.Importance = widget.DangerImportance

				// Create containers
				leftContent := container.NewHBox(
					container.NewPadded(icon),
					container.NewVBox(accountLabel, subtitleLabel),
				)

				buttons := container.NewHBox(
					layout.NewSpacer(),
					loginBtn,
					container.NewPadded(deleteBtn),
				)

				// Create main content
				mainContent := container.NewBorder(
					nil, nil,
					leftContent, buttons,
				)

				return createStyledCard(mainContent)
			},
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				if id >= len(state.Accounts) {
					return
				}

				accountName := state.Accounts[id]

				// Get the main card container
				card := obj.(*fyne.Container)
				if len(card.Objects) < 4 {
					return
				}

				// Get the content container
				borderContainer := card.Objects[3].(*fyne.Container)
				mainContent := borderContainer.Objects[0].(*fyne.Container)

				// Get the left content with labels
				leftContent := mainContent.Objects[0].(*fyne.Container)
				labelVBox := leftContent.Objects[1].(*fyne.Container)
				accountLabel := labelVBox.Objects[0].(*widget.Label)
				accountLabel.SetText(accountName)

				// Get the buttons
				buttons := mainContent.Objects[1].(*fyne.Container)
				if len(buttons.Objects) < 3 {
					return
				}

				// Get individual buttons
				loginBtn := buttons.Objects[1].(*widget.Button)
				deleteContainer := buttons.Objects[2].(*fyne.Container)
				deleteBtn := deleteContainer.Objects[0].(*widget.Button)

				// Set button handlers
				thisAccount := accountName
				loginBtn.OnTapped = func() {
					state.CurrentAccount = thisAccount
					showLoginScreen(thisAccount)()
				}

				deleteBtn.OnTapped = func() {
					state.CurrentAccount = thisAccount
					showDeleteAccountConfirmation(thisAccount)()
				}
			},
		)
	}

	// Add a solid black background
	background := canvas.NewRectangle(color.Black)

	// Load the logo image
	logo := canvas.NewImageFromFile("logo.png")
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(150, 150)) // Set a larger size for the logo

	// Styled header with shadow effect
	subHeaderText := canvas.NewText("Select an account to continue", color.White)
	subHeaderText.TextSize = 16

	header := container.NewMax(
		background,
		container.NewCenter(
			container.NewVBox(
				logo,
				subHeaderText,
			),
		),
	)

	// Styled "Add Account" button
	addAccountBtn := widget.NewButtonWithIcon("Add Account", theme.ContentAddIcon(), showAddAccountDialog)
	addAccountBtn.Importance = widget.HighImportance

	toolbar := container.NewPadded(container.NewHBox(
		layout.NewSpacer(),
		addAccountBtn,
	))

	content := container.NewBorder(
		header,
		toolbar,
		nil,
		nil,
		container.NewPadded(state.AccountList),
	)

	state.MainWindow.SetContent(content)
	refreshAccountList()

}

func showLoginScreen(accountName string) func() {
	return withErrorHandler(func() {
		log.Println("showLoginScreen called for account:", accountName) // ADDED LOG - FUNCTION ENTRY
		state.CurrentScreen = "login"
		state.CurrentAccount = accountName

		// Create styled login form
		headerText := canvas.NewText("Welcome Back", color.RGBA{R: 0x00, G: 0x88, B: 0xff, A: 0xff})
		headerText.TextSize = 48 // Increased text size
		headerText.TextStyle = fyne.TextStyle{Bold: true}
		accountText := canvas.NewText(accountName, color.White)
		accountText.TextSize = 32 // Increased text size

		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.PlaceHolder = "Enter Master Password"
		passwordEntry.Resize(fyne.NewSize(800, 60)) // Increased size

		// Create loading overlay
		loadingProgress := widget.NewProgressBarInfinite()
		loadingOverlay := container.NewVBox(
			widget.NewLabel("Verifying credentials..."),
			loadingProgress,
		)
		loadingOverlay.Hide()

		loginButton := widget.NewButton("Login", func() {
			log.Println("Login button inside showLoginScreen clicked") // ADDED LOG - BUTTON CLICK
			if passwordEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("Master password cannot be empty"), state.MainWindow)
				return
			}

			loadingOverlay.Show()

			go func() {
				entries, err := storage.LoadEncryptedData(state.CurrentAccount, passwordEntry.Text)
				if err != nil {
					log.Println("Error loading encrypted data:", err) // ADDED LOG - STORAGE ERROR
					loadingOverlay.Hide()
					dialog.ShowError(fmt.Errorf("Login failed: %w", err), state.MainWindow)
					return
				}

				state.MasterPassword = passwordEntry.Text
				state.Entries = entries

				// Refresh specific components instead of the entire canvas
				loadingOverlay.Hide()
				showDashboard()
			}()
		})
		loginButton.Importance = widget.HighImportance
		loginButton.Resize(fyne.NewSize(200, 60)) // Increased size

		backButton := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), showAccountsScreen)
		backButton.Resize(fyne.NewSize(150, 40))

		formContent := container.NewVBox(
			container.NewCenter(headerText),
			container.NewCenter(accountText),
			widget.NewLabel(""), // Spacing
			passwordEntry,
			loginButton,
			widget.NewLabel(""), // Spacing
			container.NewHBox(layout.NewSpacer(), backButton),
		)

		// Wrap in a card
		content := createStyledCard(formContent)

		// Create final layout with loading overlay
		finalContent := container.NewMax(
			content,
			container.NewCenter(loadingOverlay),
		)

		state.MainWindow.SetContent(container.NewCenter(finalContent))
		state.MainWindow.SetTitle(fmt.Sprintf("Kryptos Login - %s", accountName))

		// Focus password entry
		passwordEntry.FocusGained()
	})
}

func refreshAccountList() {
	if state.AccountList != nil {
		state.AccountList.Refresh()
	}
}

func showAddAccountDialog() {
	// Create a container for consistent padding
	paddedContainer := container.NewPadded

	// Styled header
	headerText := canvas.NewText("Create New Account", color.RGBA{R: 0x00, G: 0x88, B: 0xff, A: 0xff})
	headerText.TextSize = 20
	headerText.TextStyle = fyne.TextStyle{Bold: true}

	// Styled input fields with better spacing
	accountName := widget.NewEntry()
	accountName.SetPlaceHolder("Enter account name")
	accountName.Resize(fyne.NewSize(300, 40))

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Enter master password")
	password.Resize(fyne.NewSize(300, 40))

	confirmPassword := widget.NewPasswordEntry()
	confirmPassword.SetPlaceHolder("Confirm master password")
	confirmPassword.Resize(fyne.NewSize(300, 40))

	// Password strength indicator
	strengthBar := canvas.NewRectangle(color.Gray{Y: 0x80})
	strengthBar.Resize(fyne.NewSize(300, 10))

	strengthLabel := widget.NewLabel("Password Strength: ")
	strengthLabel.TextStyle = fyne.TextStyle{Bold: true}

	password.OnChanged = func(pass string) {
		strength := calculatePasswordStrength(pass)
		switch strength {
		case "Weak":
			strengthBar.FillColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
			strengthLabel.SetText("Password Strength: Weak")
		case "Medium":
			strengthBar.FillColor = color.RGBA{R: 255, G: 165, B: 0, A: 255}
			strengthLabel.SetText("Password Strength: Medium")
		case "Strong":
			strengthBar.FillColor = color.RGBA{R: 0, G: 255, B: 0, A: 255}
			strengthLabel.SetText("Password Strength: Strong")
		}
		strengthBar.Refresh()
	}

	// Styled form with better spacing
	form := container.NewVBox(
		paddedContainer(container.NewCenter(headerText)),
		paddedContainer(container.NewVBox(
			widget.NewLabelWithStyle("Account Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			accountName,
		)),
		paddedContainer(container.NewVBox(
			widget.NewLabelWithStyle("Master Password", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			password,
			strengthLabel,
			strengthBar,
		)),
		paddedContainer(container.NewVBox(
			widget.NewLabelWithStyle("Confirm Password", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			confirmPassword,
		)),
	)

	// Add validation and creation logic
	dialog := dialog.NewCustomConfirm(
		"Create Account",
		"Create",
		"Cancel",
		form,
		func(confirmed bool) {
			if !confirmed {
				return
			}

			// Validate inputs
			if accountName.Text == "" {
				dialog.ShowError(fmt.Errorf("Account name cannot be empty"), state.MainWindow)
				return
			}

			if len(password.Text) < 8 {
				dialog.ShowError(fmt.Errorf("Password must be at least 8 characters long"), state.MainWindow)
				return
			}

			if password.Text != confirmPassword.Text {
				dialog.ShowError(fmt.Errorf("Passwords do not match"), state.MainWindow)
				return
			}

			// Check for existing account
			for _, acc := range state.Accounts {
				if acc == accountName.Text {
					dialog.ShowError(fmt.Errorf("Account already exists"), state.MainWindow)
					return
				}
			}

			// Create account
			state.Accounts = append(state.Accounts, accountName.Text)
			if err := storage.SaveAccounts(state.Accounts); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to save account: %v", err), state.MainWindow)
				return
			}

			// Initialize account with empty entries list
			if err := storage.SaveEncryptedData(accountName.Text, password.Text, []entry.PasswordEntry{}); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to initialize account: %v", err), state.MainWindow)
				return
			}

			refreshAccountList()
			dialog.ShowInformation("Success", "Account created successfully!", state.MainWindow)
		},
		state.MainWindow,
	)

	// Style the dialog
	dialog.Resize(fyne.NewSize(400, 600))
	dialog.Show()
}

func showDeleteAccountConfirmation(accountToDelete string) func() {
	return func() {
		log.Println("showDeleteAccountConfirmation called for account:", accountToDelete) // ADDED LOG - FUNCTION ENTRY
		dialog.ShowConfirm("Delete Account", fmt.Sprintf("Are you sure you want to delete account '%s'?", accountToDelete), func(confirmed bool) {
			log.Println("Delete confirmation dialog result:", confirmed) // ADDED LOG - DIALOG RESULT
			if confirmed {
				var updatedAccounts []string
				for _, account := range state.Accounts {
					if account != accountToDelete {
						updatedAccounts = append(updatedAccounts, account)
					}
				}
				state.Accounts = updatedAccounts
				storage.SaveAccounts(state.Accounts)
				if state.CurrentAccount == accountToDelete {
					state.Entries = nil
					state.CurrentAccount = ""
				}
				refreshAccountList()
			}
		}, state.MainWindow)
	}
}

func showDashboard() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in showDashboard: %v", r)
			dialog.ShowError(fmt.Errorf("An unexpected error occurred"), state.MainWindow)
			showAccountsScreen()
		}
	}()

	// Reload entries from storage
	entries, err := storage.LoadEncryptedData(state.CurrentAccount, state.MasterPassword)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to load entries: %w", err), state.MainWindow)
		return
	}
	state.Entries = entries
	state.OriginalEntries = entries

	// Reset the entry list
	state.EntryList = nil

	state.CurrentScreen = "dashboard"
	state.MainWindow.SetTitle(fmt.Sprintf("Kryptos Dashboard - %s", state.CurrentAccount))

	// Initialize EntryList
	state.EntryList = widget.NewList(
		func() int {
			return len(state.Entries)
		},
		func() fyne.CanvasObject {
			// Title with larger, bold text
			titleLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			titleLabel.Resize(fyne.NewSize(0, 24))

			// Username row with copy button
			usernameLabel := widget.NewLabel("")
			copyUsernameBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), nil)
			copyUsernameBtn.Importance = widget.LowImportance // Makes button dark
			usernameBox := container.NewHBox(
				usernameLabel,
				copyUsernameBtn, // Moved next to username
			)

			// Password row with copy button - increased width
			passwordEntry := widget.NewPasswordEntry()
			passwordEntry.Disable()
			passwordEntry.Resize(fyne.NewSize(400, 40)) // Use Resize instead of MinSize
			copyPasswordBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), nil)
			copyPasswordBtn.Importance = widget.LowImportance // Makes button dark
			passwordBox := container.NewHBox(
				passwordEntry,
				copyPasswordBtn, // Moved next to password
			)

			// Action buttons with colors
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			editBtn.Importance = widget.HighImportance // Makes button green
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			deleteBtn.Importance = widget.DangerImportance // Makes button red
			actionButtons := container.NewHBox(
				layout.NewSpacer(),
				editBtn,
				deleteBtn,
			)

			// Main content in VBox
			content := container.NewVBox(
				titleLabel,
				usernameBox,
				passwordBox,
				widget.NewSeparator(),
				actionButtons,
			)

			// Add padding and style
			styledContent := container.NewPadded(content)
			return createStyledCard(styledContent)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(state.Entries) {
				return
			}

			entry := state.Entries[id]
			log.Printf("Updating entry: %s", entry.Title) // Debug log

			// Get the main card container
			card := obj.(*fyne.Container)
			if len(card.Objects) < 4 {
				log.Printf("Card has insufficient objects: %d", len(card.Objects))
				return
			}

			// Get the border container (4th layer in stack)
			borderContainer := card.Objects[3].(*fyne.Container)

			// Get the main content (inside border)
			mainContent := borderContainer.Objects[0].(*fyne.Container)

			// Get the VBox containing all elements
			vbox := mainContent.Objects[0].(*fyne.Container)

			// Update title
			titleLabel := vbox.Objects[0].(*widget.Label)
			titleLabel.SetText(entry.Title)

			// Update username
			usernameBox := vbox.Objects[1].(*fyne.Container)
			usernameLabel := usernameBox.Objects[0].(*widget.Label)
			usernameLabel.SetText(entry.Username)
			copyUsernameBtn := usernameBox.Objects[1].(*widget.Button)

			// Update password
			passwordBox := vbox.Objects[2].(*fyne.Container)
			passwordEntry := passwordBox.Objects[0].(*widget.Entry)
			passwordEntry.SetText(entry.Password)
			copyPasswordBtn := passwordBox.Objects[1].(*widget.Button)

			// Get action buttons
			actionButtons := vbox.Objects[4].(*fyne.Container)
			editBtn := actionButtons.Objects[1].(*widget.Button)
			deleteBtn := actionButtons.Objects[2].(*widget.Button)

			// Set button handlers with proper closure
			thisEntry := entry // Create local copy for closure

			copyUsernameBtn.OnTapped = func() {
				state.MainWindow.Clipboard().SetContent(thisEntry.Username)
				showToast("Username copied!", 2*time.Second)
			}

			copyPasswordBtn.OnTapped = func() {
				state.MainWindow.Clipboard().SetContent(thisEntry.Password)
				showToast("Password copied!", 2*time.Second)
			}

			editBtn.OnTapped = showAddEditEntryForm(&thisEntry)
			deleteBtn.OnTapped = func() {
				dialog.ShowConfirm("Delete Entry",
					fmt.Sprintf("Are you sure you want to delete '%s'?", thisEntry.Title),
					func(confirmed bool) {
						if confirmed {
							deleteEntry(thisEntry.ID)
						}
					}, state.MainWindow)
			}
		},
	)

	// Create modern search bar with icon
	searchIcon := widget.NewIcon(theme.SearchIcon())
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search passwords...")
	searchEntry.OnChanged = func(text string) {
		refreshEntryList(text)
	}

	// Create a container for the search entry
	searchContainer := container.NewBorder(nil, nil, searchIcon, nil, searchEntry)

	// Set the minimum size for the search container to ensure it stretches
	searchContainer.Resize(fyne.NewSize(600, 40))

	// Create the search box with icon
	searchBox := container.NewHBox(
		searchContainer,
		layout.NewSpacer(), // This will push the search box to expand
	)

	// Modern action buttons
	addButton := widget.NewButtonWithIcon("New Password", theme.ContentAddIcon(), showAddEditEntryForm(nil))
	addButton.Importance = widget.HighImportance

	settingsButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), showSettingsScreen)
	settingsButton.Importance = widget.MediumImportance

	// Add back button
	backButton := widget.NewButtonWithIcon("Back to Accounts", theme.LogoutIcon(), func() {
		state.CurrentAccount = ""
		state.Entries = nil
		showAccountsScreen()
	})
	backButton.Importance = widget.WarningImportance // Makes it stand out

	// Create the toolbar with the search box
	toolbar := container.NewHBox(
		backButton,
		searchBox,
		layout.NewSpacer(),
		addButton,
		settingsButton,
	)

	// Create gradient and toolbar container
	gradient := canvas.NewHorizontalGradient(
		color.RGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff},
		color.RGBA{R: 0x2d, G: 0x2d, B: 0x2d, A: 0xff},
	)

	toolbarContainer := container.NewMax(
		gradient,
		container.NewPadded(toolbar),
	)

	// Main content with padding
	mainContent := container.NewBorder(
		toolbarContainer,
		nil,
		nil,
		nil,
		container.NewPadded(state.EntryList),
	)

	state.MainWindow.SetContent(mainContent)
	refreshEntryList("")
}

// Modify the refreshEntryList function
func refreshEntryList(searchTerm string) {
	// Reload entries from storage first
	entries, err := storage.LoadEncryptedData(state.CurrentAccount, state.MasterPassword)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to reload entries: %w", err), state.MainWindow)
		return
	}

	// Update both entries and original entries
	state.Entries = entries
	state.OriginalEntries = entries

	if searchTerm != "" {
		searchTerm = strings.ToLower(searchTerm)
		var filteredEntries []entry.PasswordEntry
		for _, e := range state.OriginalEntries {
			if strings.Contains(strings.ToLower(e.Title), searchTerm) ||
				strings.Contains(strings.ToLower(e.Username), searchTerm) ||
				strings.Contains(strings.ToLower(e.Notes), searchTerm) {
				filteredEntries = append(filteredEntries, e)
			}
		}
		state.Entries = filteredEntries
	}

	if state.EntryList != nil {
		state.EntryList.Refresh()
	}
}

// Add this new function to reload entries from storage
func reloadEntries() error {
	entries, err := storage.LoadEncryptedData(state.CurrentAccount, state.MasterPassword)
	if err != nil {
		return fmt.Errorf("failed to reload entries: %w", err)
	}
	state.Entries = entries
	return nil
}

// Add this function to refresh the dashboard
func refreshDashboard() {
	if err := reloadEntries(); err != nil {
		dialog.ShowError(err, state.MainWindow)
		return
	}
	showDashboard()
}

func showAddEditEntryForm(currentEntry *entry.PasswordEntry) func() {
	return withErrorHandler(func() {
		isEdit := currentEntry != nil
		titleStr := "Add New Entry"
		if isEdit {
			titleStr = "Edit Entry"
		}

		// Create header
		header := widget.NewLabelWithStyle(titleStr, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		header.TextStyle.Bold = true

		// Create form fields
		titleEntry := widget.NewEntry()
		usernameEntry := widget.NewEntry()
		passwordEntry := widget.NewPasswordEntry()
		notesEntry := widget.NewMultiLineEntry()
		tagsEntry := widget.NewEntry()

		// Set placeholders
		titleEntry.SetPlaceHolder("Enter title")
		usernameEntry.SetPlaceHolder("Enter username")
		passwordEntry.SetPlaceHolder("Enter password")
		notesEntry.SetPlaceHolder("Enter notes (optional)")
		tagsEntry.SetPlaceHolder("Enter tags, separated by commas")

		// Password strength indicator
		strengthBar := canvas.NewRectangle(color.Gray{Y: 0x80})
		strengthBar.SetMinSize(fyne.NewSize(200, 10))

		passwordEntry.OnChanged = func(password string) {
			strength := calculatePasswordStrength(password)
			switch strength {
			case "Weak":
				strengthBar.FillColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
			case "Medium":
				strengthBar.FillColor = color.RGBA{R: 255, G: 165, B: 0, A: 255}
			case "Strong":
				strengthBar.FillColor = color.RGBA{R: 0, G: 255, B: 0, A: 255}
			}
			strengthBar.Refresh()
		}

		// Pre-fill form if editing
		if isEdit {
			titleEntry.SetText(currentEntry.Title)
			usernameEntry.SetText(currentEntry.Username)
			passwordEntry.SetText(currentEntry.Password)
			notesEntry.SetText(currentEntry.Notes)
			tagsEntry.SetText(strings.Join(currentEntry.Tags, ", "))
		}

		// Create buttons
		saveButton := widget.NewButton("Save", func() {
			title := titleEntry.Text
			username := usernameEntry.Text
			password := passwordEntry.Text
			notes := notesEntry.Text
			tags := strings.Split(tagsEntry.Text, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}

			if title == "" || username == "" || password == "" {
				dialog.ShowError(fmt.Errorf("Title, Username, and Password cannot be empty"), state.MainWindow)
				return
			}

			var updatedEntry entry.PasswordEntry
			if isEdit {
				updatedEntry = *currentEntry
				updatedEntry.Title = title
				updatedEntry.Username = username
				updatedEntry.Password = password
				updatedEntry.Notes = notes
				updatedEntry.Tags = tags
				updatedEntry.UpdatedAt = time.Now()

				var newEntries []entry.PasswordEntry
				for _, e := range state.Entries {
					if e.ID == updatedEntry.ID {
						newEntries = append(newEntries, updatedEntry)
					} else {
						newEntries = append(newEntries, e)
					}
				}
				state.Entries = newEntries
			} else {
				updatedEntry = entry.NewPasswordEntry(title, username, password, "", notes, tags)
				state.Entries = append(state.Entries, updatedEntry)
			}

			// Save to storage first
			err := storage.SaveEncryptedData(state.CurrentAccount, state.MasterPassword, state.Entries)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to save entry: %w", err), state.MainWindow)
				return
			}

			// Update both state entries and original entries
			state.OriginalEntries = state.Entries

			// Force recreation of the dashboard
			showDashboard()
		})
		saveButton.Importance = widget.HighImportance

		cancelButton := widget.NewButton("Cancel", showDashboard)

		// Create form items
		form := widget.NewForm(
			widget.NewFormItem("Title", titleEntry),
			widget.NewFormItem("Username", usernameEntry),
			widget.NewFormItem("Password", container.NewVBox(
				passwordEntry,
				container.NewHBox(
					widget.NewLabel("Strength:"),
					strengthBar,
				),
			)),
			widget.NewFormItem("Notes", notesEntry),
			widget.NewFormItem("Tags", tagsEntry),
		)

		// Create buttons container
		buttons := container.NewHBox(
			layout.NewSpacer(),
			cancelButton,
			saveButton,
		)

		// Create main content
		content := container.NewVBox(
			header,
			form,
			buttons,
		)

		// Wrap in a card with padding
		card := createStyledCard(container.NewPadded(content))

		state.MainWindow.SetContent(container.NewCenter(card))
		state.MainWindow.SetTitle(titleStr)
	})
}

func calculatePasswordStrength(password string) string {
	// Simple example logic for password strength
	if len(password) < 6 {
		return "Weak"
	} else if len(password) < 10 {
		return "Medium"
	}
	return "Strong"
}

func showDeleteConfirmation(entryToDelete *entry.PasswordEntry) func() {
	return withErrorHandler(func() {
		dialog.ShowConfirm("Delete Entry", fmt.Sprintf("Are you sure you want to delete entry '%s'?", entryToDelete.Title), func(confirmed bool) {
			if confirmed {
				var updatedEntries []entry.PasswordEntry
				for _, entry := range state.Entries {
					if entry.ID != entryToDelete.ID {
						updatedEntries = append(updatedEntries, entry)
					}
				}
				state.Entries = updatedEntries
				err := storage.SaveEncryptedData(state.CurrentAccount, state.MasterPassword, state.Entries)
				if err != nil {
					dialog.ShowError(fmt.Errorf("Failed to delete entry: %w", err), state.MainWindow)
				}
				refreshEntryList("") // Refresh list after deletion
			}
		}, state.MainWindow)
	})
}

func showSettingsScreen() {
	withErrorHandler(func() {
		state.MainWindow.SetTitle(fmt.Sprintf("Settings - %s", state.CurrentAccount))

		exportButton := widget.NewButton("Export Passwords", showExportDialog)
		importButton := widget.NewButton("Import Passwords", showImportDialog)
		backButton := widget.NewButton("<- Back to Dashboard", showDashboard)

		content := container.NewVBox(
			widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			layout.NewSpacer(),
			widget.NewLabel("Export & Import"),
			exportButton,
			importButton,
			layout.NewSpacer(),
			backButton,
		)
		state.MainWindow.SetContent(container.NewCenter(content))
	})()
}

func showExportDialog() {
	dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, state.MainWindow)
			return
		}
		if writer == nil { // User cancelled
			return
		}

		defer writer.Close()

		err = storage.ExportEncryptedData(state.CurrentAccount, state.MasterPassword, writer.URI().Path())
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to export data: %w", err), state.MainWindow)
		} else {
			dialog.ShowInformation("Export Successful", "Passwords exported successfully!", state.MainWindow)
		}

	}, state.MainWindow).Show()
}

func showImportDialog() {
	dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, state.MainWindow)
			return
		}
		if reader == nil { // User cancelled
			return
		}
		defer reader.Close()

		err = storage.ImportEncryptedData(state.CurrentAccount, state.MasterPassword, reader.URI().Path())
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to import data: %w", err), state.MainWindow)
		} else {
			// After import, reload entries and refresh dashboard
			entries, err := storage.LoadEncryptedData(state.CurrentAccount, state.MasterPassword)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to load imported data: %w", err), state.MainWindow)
				return
			}
			state.Entries = entries
			refreshEntryList("")
			showDashboard() // Go back to dashboard
			dialog.ShowInformation("Import Successful", "Passwords imported successfully!", state.MainWindow)
		}
	}, state.MainWindow).Show()
}

// DarkTheme implementation (you can refine this further)
type darkTheme struct{}

var _ fyne.Theme = (*darkTheme)(nil)

func (d darkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return theme.DefaultTheme().Color(name, variant)
	}
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff} // Dark grey background
	case theme.ColorNameForeground:
		return color.White // White text
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x64, G: 0xb5, B: 0xf6, A: 0xff} // Example primary color (blue)
	case theme.ColorNameHover:
		return color.RGBA{R: 0x50, G: 0x50, B: 0x50, A: 0xff} // Slightly lighter background on hover
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xff} // Darker input background
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (d darkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style) // Use default font for simplicity
}

func (d darkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name) // Use default icons
}

func (d darkTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name) // Use default sizes
}

// ApplyDarkTheme sets the dark theme for the application.
func ApplyDarkTheme(a fyne.App) {
	a.Settings().SetTheme(darkTheme{})
}

func showLoadingScreen(message string) func() {
	return func() {
		progress := widget.NewProgressBarInfinite()
		label := widget.NewLabel(message)

		content := container.NewVBox(
			label,
			progress,
		)

		dialog := dialog.NewCustom("Loading", "Cancel", content, state.MainWindow)
		dialog.Show()
	}
}

// Custom theme with vibrant colors and stylish fonts
type vibrantTheme struct{}

func (v vibrantTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff} // Darker background
	case theme.ColorNameForeground:
		return color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // White text
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x00, G: 0xff, B: 0x88, A: 0xff} // Vibrant green
	case theme.ColorNameButton:
		return color.RGBA{R: 0x00, G: 0x88, B: 0xff, A: 0xff} // Bright blue
	case theme.ColorNameHover:
		return color.RGBA{R: 0x3d, G: 0x3d, B: 0x3d, A: 0xff} // Lighter on hover
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (v vibrantTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (v vibrantTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (v vibrantTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14 // Larger text size
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func ApplyVibrantTheme(a fyne.App) {
	a.Settings().SetTheme(vibrantTheme{})
}

// Simplify createStyledCard function
func createStyledCard(content fyne.CanvasObject) fyne.CanvasObject {
	// Create main background
	bg := canvas.NewRectangle(cardBgColor)

	// Create golden accent line
	accentLine := canvas.NewRectangle(royalGold)
	accentLine.Resize(fyne.NewSize(4, 140))

	// Create border with subtle glow
	border := canvas.NewRectangle(color.RGBA{R: 0x28, G: 0x28, B: 0x28, A: 0xff})

	// Add subtle shadow
	shadow := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 0x40})
	shadow.Move(fyne.NewPos(2, 2))

	// Create the base card with all layers
	card := container.NewMax(
		shadow,
		border,
		bg,
		container.NewBorder(
			nil, nil,
			accentLine,
			nil,
			content,
		),
	)

	return card
}

// Add this function to create styled buttons
func createStyledButton(text string, icon fyne.Resource, onTapped func()) *widget.Button {
	btn := widget.NewButtonWithIcon(text, icon, onTapped)
	btn.Importance = widget.MediumImportance

	// Custom styling
	btn.Resize(fyne.NewSize(120, 40))

	return btn
}

// Add this function to create styled entry fields
func createStyledEntry(placeholder string) *widget.Entry {
	entry := widget.NewEntry()
	entry.PlaceHolder = placeholder
	entry.TextStyle = fyne.TextStyle{
		Bold: false,
	}

	return entry
}

// Helper function to show temporary toast messages
func showToast(message string, duration time.Duration) {
	toast := widget.NewLabel(message)
	toast.Alignment = fyne.TextAlignCenter

	bg := canvas.NewRectangle(color.RGBA{R: 0x2d, G: 0x2d, B: 0x2d, A: 0xee})

	content := container.NewMax(bg, container.NewPadded(toast))
	content.Resize(fyne.NewSize(200, 40))

	// Create a new window for the toast
	toastWindow := state.App.NewWindow("")
	toastWindow.SetContent(content)
	toastWindow.Resize(fyne.NewSize(200, 40))

	// Position the toast window relative to main window
	// mainSize := state.MainWindow.Content().Size()
	toastWindow.CenterOnScreen() // Center the toast window

	toastWindow.Show()

	// Hide after duration
	go func() {
		time.Sleep(duration)
		toastWindow.Close()
	}()
}

// Add animations to buttons
func animateButton(btn *widget.Button) {
	red := color.NRGBA{R: 0xff, A: 0xff}
	blue := color.NRGBA{B: 0xff, A: 0xff}
	canvas.NewColorRGBAAnimation(red, blue, time.Second*2, func(c color.Color) {
		btn.Importance = widget.HighImportance
		btn.Refresh()
	}).Start()
}

// Add this general error handler wrapper function
func withErrorHandler(action func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v", r)
				dialog.ShowError(fmt.Errorf("An unexpected error occurred. Please try again."), state.MainWindow)
			}
		}()
		action()
	}
}

// Add this helper function for safe type assertions
func safeGet(obj fyne.CanvasObject, indices ...int) (fyne.CanvasObject, bool) {
	current := obj
	for _, i := range indices {
		if container, ok := current.(*fyne.Container); ok && i < len(container.Objects) {
			current = container.Objects[i]
		} else {
			return nil, false
		}
	}
	return current, true
}

// Helper function for creating styled entry cards with gradient
func createStyledEntryCard(content fyne.CanvasObject) *fyne.Container {
	// Create background container with gradient
	background := canvas.NewRectangle(color.RGBA{R: 0x2d, G: 0x2d, B: 0x2d, A: 0xff})

	// Add gradient overlay
	gradient := canvas.NewLinearGradient(
		color.RGBA{R: 0x2d, G: 0x2d, B: 0x2d, A: 0xff},
		color.RGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff},
		0,
	)
	gradient.Resize(background.Size())

	// Add shadow
	shadow := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 0x33})
	shadow.Move(fyne.NewPos(2, 2))

	// Create the main container with all layers
	card := container.NewMax(
		shadow,
		background,
		gradient,
		container.NewPadded(content),
	)

	// Set minimum size for consistency
	background.Resize(fyne.NewSize(0, 140))
	gradient.Resize(fyne.NewSize(0, 140))
	shadow.Resize(fyne.NewSize(0, 140))

	return card
}

// Modify the deleteEntry function
func deleteEntry(entryID string) {
	var updatedEntries []entry.PasswordEntry
	for _, e := range state.Entries {
		if e.ID != entryID {
			updatedEntries = append(updatedEntries, e)
		}
	}

	// Save the updated entries
	err := storage.SaveEncryptedData(state.CurrentAccount, state.MasterPassword, updatedEntries)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to delete entry: %v", err), state.MainWindow)
		return
	}

	// Update both state entries and original entries
	state.Entries = updatedEntries
	state.OriginalEntries = updatedEntries

	// Force recreation of the dashboard
	showDashboard()
	showToast("Entry deleted successfully!", 2*time.Second)
}
