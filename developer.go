package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Team struct {
	Name     string `json:"name"`
	Score    int    `json:"score"`
	Attempts int    `json:"attempts"`
	Password string `json:"password"` // Add this field
}

type Riddle struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

var firebaseApp *firebase.App

os.Getenv("FIREBASE_CREDENTIALS")

func initFirebase() {
    firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS")
    if firebaseCredentials == "" {
        log.Fatal("FIREBASE_CREDENTIALS environment variable is not set")
    }
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

	passwordQuery := client.Collection("passwords").Limit(1).Documents(ctx)
	docs, err := passwordQuery.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching password from Firebase: %v", err)
	}
	if len(docs) == 0 {
		return fmt.Errorf("no password found in Firebase")
	}

	doc := docs[0]
	var data map[string]string
	if err := doc.DataTo(&data); err != nil {
		return fmt.Errorf("error reading password data: %v", err)
	}

	storedPassword := data["password"]
	if storedPassword != currentPassword {
		return fmt.Errorf("current password does not match")
	}

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
	red := color.New(color.FgHiRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
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
		err = doc.DataTo(&team)
		if err != nil {
			log.Fatalf("Error converting document data to Team struct: %v\n", err)
		}

		// Display the team details, including the password
		fmt.Printf(green("Team:")+" %s"+green(",\tScore:")+" %d"+green(",\tAttempts:")+" %d"+green(",")+red("\tPassword:")+" %s\n", team.Name, team.Score, team.Attempts, team.Password)
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

func addApprovedTeamToFirebase(teamName string) {
	blue := color.New(color.FgBlue).SprintFunc()
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	// Add team name to the approved_teams collection
	_, err = client.Collection("approved_teams").Doc(teamName).Set(ctx, map[string]interface{}{
		"name": teamName,
	})
	if err != nil {
		log.Fatalf("Error adding approved team to Firebase: %v\n", err)
	}
	fmt.Printf(blue("Approved team")+" %s "+(blue("added successfully!\n\n")), teamName)
}

func setGameDurationInFirebase(duration int) error {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("error creating Firestore client: %v", err)
	}
	defer client.Close()

	_, err = client.Collection("game_settings").Doc("duration").Set(ctx, map[string]interface{}{
		"minutes": duration,
	})
	if err != nil {
		return fmt.Errorf("error setting game duration in Firebase: %v", err)
	}

	return nil
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

func deleteAllRiddlesFromFirebase() error {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("error creating Firestore client: %v", err)
	}
	defer client.Close()

	// Get all documents in the "riddles" collection
	iter := client.Collection("riddles").Documents(ctx)
	batch := client.Batch()

	// Add delete operations to the batch
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating through riddles: %v", err)
		}
		batch.Delete(doc.Ref)
	}

	// Commit the batch
	_, err = batch.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error deleting riddles: %v", err)
	}

	return nil
}

func viewApprovedTeamsInFirebase() {
	blue := color.New(color.FgBlue).SprintFunc()
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	iter := client.Collection("approved_teams").Documents(ctx)
	fmt.Println(blue("\nApproved Teams:-"))
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating through approved teams: %v\n", err)
		}
		var data map[string]interface{}
		doc.DataTo(&data)
		fmt.Println(data["name"])
	}
	println()
}

func viewRiddlesInFirebase() {
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	iter := client.Collection("riddles").Documents(ctx)
	fmt.Println(blue("\nAll Riddles:"))
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating through riddles: %v\n", err)
		}
		var riddle Riddle
		doc.DataTo(&riddle)
		fmt.Printf(green("Question: ")+"%s\n"+green("Answer: ")+"%s\n\n", riddle.Question, riddle.Answer)
	}
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
		fmt.Println("3. Change admin Password")
		fmt.Println("4. Add Approved Team")
		fmt.Println("5. View Approved Teams")
		fmt.Println("6. Set Game Duration")
		fmt.Println("7. Delete All Riddles")
		fmt.Println("8. View All Riddles") // New option
		fmt.Println("9. Exit")
		fmt.Print(green("Choose an option: "))

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			viewTeamsInFirebase()
			fmt.Println()
		case 2:
			fmt.Print(green("Enter the riddle question: "))
			question, _ := reader.ReadString('\n')
			question = strings.TrimSpace(question)

			fmt.Print(green("Enter the riddle answer: "))
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(answer)

			riddle := Riddle{
				Question: question,
				Answer:   answer,
			}

			addRiddleToFirebase(riddle)
			fmt.Println(blue("Riddle added successfully!\n"))
		case 3:
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
			fmt.Print(green("Enter the team name to approve: "))
			teamName, _ := reader.ReadString('\n')
			teamName = strings.TrimSpace(teamName)

			addApprovedTeamToFirebase(teamName)
		case 5:
			viewApprovedTeamsInFirebase()
		case 6:
			fmt.Print(green("Enter the game duration in minutes: "))
			durationStr, _ := reader.ReadString('\n')
			durationStr = strings.TrimSpace(durationStr)
			duration, err := strconv.Atoi(durationStr)
			if err != nil {
				fmt.Println(red("Invalid input. Please enter a number."))
				continue
			}
			err = setGameDurationInFirebase(duration)
			if err != nil {
				fmt.Printf(red("Error setting game duration: %v\n", err))
			} else {
				fmt.Println(blue("Game duration set successfully!\n"))
			}
		case 7:
			fmt.Print(red("Are you sure you want to delete all riddles? This action cannot be undone. (y/n): "))
			confirmation, _ := reader.ReadString('\n')
			confirmation = strings.TrimSpace(strings.ToLower(confirmation))

			if confirmation == "y" {
				err := deleteAllRiddlesFromFirebase()
				if err != nil {
					fmt.Printf(red("Error deleting riddles: %v\n", err))
				} else {
					fmt.Println(blue("All riddles deleted successfully!\n"))
				}
			} else {
				fmt.Println(blue("Riddle deletion cancelled.\n"))
			}
		case 8:
			viewRiddlesInFirebase() // New case to view all riddles
		case 9:
			fmt.Println(blue("Exiting..."))
			return
		default:
			fmt.Println(red("Invalid option"))
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	initFirebase()       // Initialize Firebase
	developerInterface() // Run developer interface
}
