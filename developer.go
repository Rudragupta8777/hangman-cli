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

const firebaseCredentials = `{
	"type": "service_account",
	"project_id": "hangman-cli",
	"private_key_id": "2966177952a3d009b6ab832ba462f0f02ae2f73c",
	"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDb+H48dkhe90ld\nNvkWLevaIhzEBQoWH1hOBWNjkv6cUmbV9D/mH9ZsocFZsftBLDG4vPXVcjwrEMKC\n28fE8gdLSD/W1VOB9YNFl2wlVoHRg2SrzE56jGNcWRHQgz+R/kh+4yFBiuKmTcX2\n4Zh91z5iXMT7XnyD4FLyaBw8j4urjFcdG1PUJC9cEugH08RXuPFe7ckZaZDu+lfA\ndp8P1JSue5Z9lLeO6j0r91a7p7z4GTUEYE8zVtYRN5kFS1OQ5AcyI5WxpHwidAVJ\nBtMalDYgPQQKanwrnHCWClMH2AypeN3qigkLPhwDdqKpWuuoE5d7NoxjRFl7Ku/A\no6xe60zvAgMBAAECggEAUy8oEdZLLPqH+GOzE4OfJtjyluAu/cmxu6OG/99VQKla\nsTtSNMTCckdDVpebY/yB+xIeRy8ReNm4LQNPCvfZ8UqrtaLrlwBQua73GzGZGzF8\njwlOfkJ7yq72MSuJDT0jjjR3XZFXf7t2ixOp9qDAuzLI3SRQoxBgXcIoN3CzSVYv\n+EOxFXVBXJChxlluXLD/dq4vH+geKPMf8blDWuTxwJyBdGtQWvIx/nJeBzAYiC/+\nw5OVb2h0KLDZPCplpM1SLBURBqM+cJPi9kwPPqX9Tv9yErTcJPd9ynCC4ZUXvW4W\n8O6eioh3UPWjTIdtTiunWklYUtp485OsFfMPCtE/8QKBgQD2pnX+4UEuZv4GQtEt\nDkIZszDLxsP4M3nG9KqwGvgZFRlejcX+Pv4MfP955fMukE9BmQO2bwsKvJl1NVpP\nKtOSoKcCW7EkflS1z/tjAF/T+yjl6C6hB7Dj06yHoas/FPIA8EMQhrAoB7zx1CKW\nkxGdlD0r6+Gb4SIcpJe3karm8QKBgQDkTyD7a/7+RgdE9AMdltlJJKnjIa7IslD+\nkZksdircdoja6mcGDq5nuPVObV3AHO0GZ3gVUxdxo25vH8lqgNathU+W8u38WJhb\nENSicir2ntrkTX6HbVMxYre60cL+wNjaxFS+0N5RWMn+PXZqANxMnFzJTZb1QZug\nChutglgx3wKBgQCJLUFY1SysQwmqr8Soe1KV+ov7+XsKco6a8X5w3T74rDxk0xK3\n+Y7PoUFxKUvbrNT3lcNz1kRc31G110t31ki/NuxLqnVV55DzYU3d3NpvCjPP0hcE\n5kMiIprFAEw+lEaX8QhLi60zRkJ2eNYXyom0izqOT+01Bbw0E/JxXOmg8QKBgA90\nG7No9/WWH9/W9G8ISuTcinNJUF9dUoYorMmJphUOIO1QeHC8hamXp2MLnBDo5FJO\npp4q5adXfJ4g9K0001MjduOsxdcS2B0x4nKsb6QJ1J8nb60TBVKOcAlBMYW03/jO\n2T2hPasb63A+EMnUDRVScCVgDxvCuRn4FS+FZxrZAoGAffwED3ER8anS6ngu1iu+\njEW/GvMM7ONb59QMlHttrQotnJ+ENXIy3wFHmnwDkBkQqQkRhGukkOJkRJ9S92Xc\nEyDRM39rh7A1zs0nKFuSiro20d2TeinLQXLZT5elq6wknjqxOkLfYFcSkvhzT1IB\n1Y6VDCbgEbJcT2l5N//OcuI=\n-----END PRIVATE KEY-----\n",
	"client_email": "firebase-adminsdk-vh83s@hangman-cli.iam.gserviceaccount.com",
	"client_id": "117632128327902282897",
	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	"token_uri": "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-vh83s%40hangman-cli.iam.gserviceaccount.com",
	"universe_domain": "googleapis.com"
}` // copy paste the firebase credientials here

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
	initFirebase()       // Initialize Firebase
	developerInterface() // Run developer interface
}
