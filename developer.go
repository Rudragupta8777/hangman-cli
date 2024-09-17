package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Team struct {
	Name      string `json:"name"`
	Score     int    `json:"score"`
	MachineID string `json:"machineID"`
	Attempts  int    `json:"attempts"`
}

type Riddle struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

var firebaseApp *firebase.App

const firebaseCredentials = `{}` // copy paste the firebase credientials here

func initFirebase() {
	opt := option.WithCredentialsJSON([]byte(firebaseCredentials))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing app: %v\n", err)
	}
	firebaseApp = app
}

func changePasswordInFirebase(currentPassword, newPassword string) error {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("error creating Firestore client: %v", err)
	}
	defer client.Close()

	// Fetch the current password
	passwordQuery := client.Collection("passwords").Limit(1).Documents(ctx)
	docs, err := passwordQuery.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching password from Firebase: %v", err)
	}
	if len(docs) == 0 {
		return fmt.Errorf("no password found in Firebase")
	}

	// Check if the current password matches
	doc := docs[0]
	var data map[string]string
	if err := doc.DataTo(&data); err != nil {
		return fmt.Errorf("error reading password data: %v", err)
	}

	storedPassword := data["password"]
	if storedPassword != currentPassword {
		return fmt.Errorf("current password does not match")
	}

	// Update the password
	_, err = client.Collection("passwords").Doc(doc.Ref.ID).Set(ctx, map[string]string{"password": newPassword})
	if err != nil {
		return fmt.Errorf("error updating password in Firebase: %v", err)
	}

	return nil
}

func saveTeamToFirebase(team Team) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	_, err = client.Collection("teams").Doc(team.Name).Set(ctx, team)
	if err != nil {
		log.Fatalf("Error writing to Firebase: %v\n", err)
	}
}

func viewTeamsInFirebase() {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	iter := client.Collection("teams").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating through teams: %v\n", err)
		}
		var team Team
		doc.DataTo(&team)
		fmt.Printf("Team: %s,\tScore: %d,\tMachineID: %s,\tAttempts: %d\n", team.Name, team.Score, team.MachineID, team.Attempts)
	}
}

func addRiddleToFirebase(riddle Riddle) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	_, _, err = client.Collection("riddles").Add(ctx, riddle)
	if err != nil {
		log.Fatalf("Error adding riddle to Firebase: %v\n", err)
	}
}

func displayLogo() {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Println(`
       
	       ██╗ ███████╗ ████████╗ ███████╗
	       ██║ ██╔════╝ ╚══██╔══╝ ██╔════╝
	       ██║ ███████╗    ██║    █████╗  
	       ██║ ╚════██║    ██║    ██╔══╝  
	       ██║ ███████║    ██║    ███████╗
	       ╚═╝ ╚══════╝    ╚═╝    ╚══════╝
	`)
	fmt.Println(yellow("\tWelcome to the Solaris Hangman Developer side!\n"))
}

func developerInterface() {
	reader := bufio.NewReader(os.Stdin)
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()

	displayLogo()

	for {
		fmt.Println(blue("  Developer CLI\n"))
		fmt.Println("1. View Teams")
		fmt.Println("2. Add Riddle")
		fmt.Println("3. Change Password")
		fmt.Println("4. Exit")
		fmt.Print(green("Choose an option: "))

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			viewTeamsInFirebase()
			fmt.Println()
		case 2:
			// Add a new riddle
			fmt.Print(green("Enter the riddle question: "))
			question, _ := reader.ReadString('\n')
			question = strings.TrimSpace(question)

			fmt.Print(green("Enter the riddle answer: "))
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			riddle := Riddle{
				Question: question,
				Answer:   answer,
			}

			addRiddleToFirebase(riddle)
			fmt.Println(blue("Riddle added successfully!\n"))
		case 3:
			// Change the password
			fmt.Print(green("Enter the current password: "))
			currentPassword, _ := reader.ReadString('\n')
			currentPassword = strings.TrimSpace(currentPassword)

			fmt.Print(green("Enter the new password: "))
			newPassword, _ := reader.ReadString('\n')
			newPassword = strings.TrimSpace(newPassword)

			err := changePasswordInFirebase(currentPassword, newPassword)
			if err != nil {
				fmt.Printf(red("Error changing password: %v\n", err) + "\n\n")
			} else {
				fmt.Println(blue("Password changed successfully!\n"))
			}
		case 4:
			fmt.Println(blue("Exiting..."))
			return
		default:
			fmt.Println(red("Invalid option"))
		}
	}
}

func main() {
	initFirebase()       // Initialize Firebase
	developerInterface() // Run developer interface
}
