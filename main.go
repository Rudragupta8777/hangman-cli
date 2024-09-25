package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/fatih/color"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type Team struct {
	Name     string `json:"name"`
	Score    int    `json:"score"`
	Attempts int    `json:"attempts"`
	Password string `json:"password"` // New field for team password
}

type Riddle struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

var riddles = []Riddle{
	// {"I'm light as a feather, yet the strongest person can't hold me for five minutes. What am I?", "breath"},
	// {"I'm found in socks, scarves and mittens; and often in the paws of playful kittens. What am I?", "yarn"},
	// {"Where does today come before yesterday?", "dictionary"},
	// {"What invention lets you look right through a wall?", "window"},
	// {"If you have me, you want to share me. If you share me, you don't have me. What am I?", "secret"},
	// {"What goes up but never comes down?", "age"},
	// {"The more you take, the more you leave behind. What am I?", "footsteps"},
	// {"What can travel around the world while staying in one corner?", "stamp"},
}

var hangmanStages = []string{
	`
	 -----
	 |   |
	     |
	     |
	     |
	     |
	 =========`,
	`
	 -----
	 |   |
	 O   |
	     |
	     |
	     |
	 =========`,
	`
	 -----
	 |   |
	 O   |
	 |   |
	     |
	     |
	 =========`,
	`
	 -----
	 |   |
	 O   |
	/|   |
	     |
	     |
	 =========`,
	`
	 -----
	 |   |
	 O   |
	/|\  |
	     |
	     |
	 =========`,
	`
	 -----
	 |   |
	 O   |
	/|\  |
	/    |
	     |
	 =========`,
	`
	 -----
	 |   |
	 O   |
	/|\  |
	/ \  |
	     |
	 =========`,
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

func getPasswordFromFirebase() (string, error) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return "", fmt.Errorf("error creating Firestore client: %v", err)
	}
	defer client.Close()

	doc, err := client.Collection("passwords").Doc("admin").Get(ctx)
	if err != nil {
		return "", fmt.Errorf("error retrieving password document: %v", err)
	}

	var data map[string]string
	doc.DataTo(&data)
	password, ok := data["password"]
	if !ok {
		return "", fmt.Errorf("password field not found in document")
	}

	return password, nil
}

func saveTeamToFirebase(team Team) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	_, err = client.Collection("teams").Doc(team.Name).Set(ctx, map[string]interface{}{
		"score":    team.Score,
		"name":     team.Name,
		"attempts": team.Attempts,
		"password": team.Password, // Save password to Firestore
	}, firestore.MergeAll)

	if err != nil {
		log.Fatalf("Error updating team in Firebase: %v\n", err)
	}
}

func getTeamFromFirebase(teamName string) (Team, error) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return Team{}, fmt.Errorf("error creating Firestore client: %v", err)
	}
	defer client.Close()

	doc, err := client.Collection("teams").Doc(teamName).Get(ctx)
	if err != nil {
		return Team{}, fmt.Errorf("error retrieving team document: %v", err)
	}

	var team Team
	err = doc.DataTo(&team)
	if err != nil {
		return Team{}, fmt.Errorf("error parsing team data: %v", err)
	}

	return team, nil
}

func getApprovedTeamsFromFirebase() ([]string, error) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating Firestore client: %v", err)
	}
	defer client.Close()

	docs, err := client.Collection("approved_teams").Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error retrieving approved teams: %v", err)
	}

	var approvedTeams []string
	for _, doc := range docs {
		approvedTeams = append(approvedTeams, doc.Ref.ID)
	}

	return approvedTeams, nil
}

func validateTeam(approvedTeams []string, teamName string) bool {
	for _, approvedTeam := range approvedTeams {
		if approvedTeam == teamName {
			return true
		}
	}
	return false
}

func createPasswordForNewTeam(team *Team, reader *bufio.Reader) {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Print(green("This is your first login. Please create a password: "))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	team.Password = password
	saveTeamToFirebase(*team)
}

func validatePassword(team *Team, reader *bufio.Reader) bool {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Print(green("Enter your password: "))
	passwordEntered, _ := reader.ReadString('\n')
	passwordEntered = strings.TrimSpace(passwordEntered)
	return passwordEntered == team.Password
}

var timeUp = make(chan bool)

func startTimer(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		timeUp <- true
	}()
}

func getGameDurationFromFirebase() (time.Duration, error) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return 0, fmt.Errorf("error creating Firestore client: %v", err)
	}
	defer client.Close()

	doc, err := client.Collection("game_settings").Doc("duration").Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("error retrieving game duration: %v", err)
	}

	var data map[string]interface{}
	doc.DataTo(&data)
	minutes, ok := data["minutes"].(int64)
	if !ok {
		return 0, fmt.Errorf("invalid game duration format")
	}

	return time.Duration(minutes) * time.Minute, nil
}

func getRiddlesFromFirebase() ([]Riddle, error) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating Firestore client: %v", err)
	}
	defer client.Close()

	var riddlesFromFirebase []Riddle
	docs, err := client.Collection("riddles").Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error retrieving riddles: %v", err)
	}

	for _, doc := range docs {
		var riddle Riddle
		err := doc.DataTo(&riddle)
		if err != nil {
			return nil, fmt.Errorf("error converting document data to riddle: %v", err)
		}
		riddlesFromFirebase = append(riddlesFromFirebase, riddle)
	}

	return riddlesFromFirebase, nil
}

func randomRiddles(num int) ([]Riddle, error) {
	firebaseRiddles, err := getRiddlesFromFirebase() // Fetch riddles from Firebase
	if err != nil {
		return nil, err
	}

	// Combine hardcoded riddles with Firebase riddles
	allRiddles := append(riddles, firebaseRiddles...)

	// Shuffle the combined list and pick the desired number of riddles
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allRiddles), func(i, j int) {
		allRiddles[i], allRiddles[j] = allRiddles[j], allRiddles[i]
	})

	// Ensure we don't try to select more riddles than exist
	if num > len(allRiddles) {
		num = len(allRiddles)
	}

	return allRiddles[:num], nil
}

func drawHangman(stage int) {
	fmt.Println(hangmanStages[stage])
}

func startScoreUpdater(team *Team) {
	go func() {
		for {
			time.Sleep(1 * time.Second) // Update score every 1 second
			saveTeamToFirebase(*team)
		}
	}()
}

func displaysolarisLogo() {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Println(yellow(`  
	                            +=                                                              
                                   #+=                                                              
                                  +#==                                                              
                                 ==-*+=                                                             
                                 ==:=*+=                                                            
         +=                      ==---*+-                                                           
         ==---                   +=-::===                                                           
         ===+=====-:-   *++   **+*#===+--                                                           
          ===-++**+===-=+*+  =****==+=-:++==--                                                      
            =-=====-=+:=--*=--+=*++=:::*#=--                                                        
              ===++-=---=====-=---::.:-==:::                                                        
                =+==+---==--=--=+==--:==:.                                                          
                =+=+=----***+==----=======                                                          
                ++++-.:+##*=-:..:  
           ==    =++.:=+===...    
           +=+==:--:.=--+=::      ███████╗  ██████╗  ██╗       █████╗  ██████╗  ██╗ ███████╗
            ===+=+====:===:       ██╔════╝ ██╔═══██╗ ██║      ██╔══██╗ ██╔══██╗ ██║ ██╔════╝
        +*+=-:--==:-===+-:        ███████╗ ██║   ██║ ██║      ███████║ ██████╔╝ ██║ ███████╗
      ==*+**++----===*++--        ╚════██║ ██║   ██║ ██║      ██╔══██║ ██╔══██╗ ██║ ╚════██║
   **===-=++++-:*====+++-=        ███████║ ╚██████╔╝ ███████╗ ██║  ██║ ██║  ██║ ██║ ███████║
=====-=***=-=+*++======+---       ╚══════╝  ╚═════╝  ╚══════╝ ╚═╝  ╚═╝ ╚═╝  ╚═╝ ╚═╝ ╚══════╝ 
  ---===:.: ---=+--===-++--- 
               :==::---=++====-==+=         
                  =+==---=++++=--===--==     
               =+===--:==:-=++++==------++*+=                                                       
               =-=-:-+*+#=:::-------=*+++#+--                                                       
                  +*##*+=:*=----=+=-=*#*=-=-:                                                       
                 ==--=+++++-:..===-:=+-==-::                                                        
                 +++++=-::...  =-+:-=*+=::.                                                         
                 =-++:..       ==-:===+=:---                                                        
                 ===+..             --*=-----                                                       
                 ==+-.               :--+*---                                                       
                ===-:                   -=+=-                                                       
                                          ==:                                                       
                                           =                                                        
	`))
	fmt.Println(yellow("\t\tWelcome to the Solaris Hangman Game!\n"))
}

func displaygameoverLogo() {
	red := color.New(color.FgHiRed).SprintFunc()
	fmt.Println(red(`

 ██████╗  █████╗ ███╗   ███╗███████╗     ██████╗ ██╗   ██╗███████╗██████╗ 
██╔════╝ ██╔══██╗████╗ ████║██╔════╝    ██╔═══██╗██║   ██║██╔════╝██╔══██╗
██║  ███╗███████║██╔████╔██║█████╗      ██║   ██║██║   ██║█████╗  ██████╔╝
██║   ██║██╔══██║██║╚██╔╝██║██╔══╝      ██║   ██║╚██╗ ██╔╝██╔══╝  ██╔══██╗
╚██████╔╝██║  ██║██║ ╚═╝ ██║███████╗    ╚██████╔╝ ╚████╔╝ ███████╗██║  ██║
 ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝     ╚═════╝   ╚═══╝  ╚══════╝╚═╝  ╚═╝
	`))
}

func runRiddles(team *Team, riddlesSubset []Riddle, reader *bufio.Reader) {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	// yellow := color.New(color.FgYellow).SprintFunc()

	wrongGuesses := 0
	for i, riddle := range riddlesSubset {
		// Check if time is up before proceeding to the next riddle
		select {
		case <-timeUp:
			fmt.Println(red("\nTime's up! The game is over."))
			displaygameoverLogo()
			return
		default:
			// Proceed with the next riddle if time is not up
		}

		if wrongGuesses >= len(hangmanStages)-1 {
			fmt.Println(red("You've been hanged!"))
			displaygameoverLogo()
			break
		}

		fmt.Printf("\n%s %s\n", green("Question "+fmt.Sprintf("%d:", i+1)), riddle.Question)
		fmt.Print(green("Enter your guess [whole word]: "))
		guess, _ := reader.ReadString('\n')
		guess = strings.TrimSpace(strings.ToLower(guess)) // No lowercase conversion

		// Case-sensitive comparison for riddle answer
		if guess == riddle.Answer {
			fmt.Println(blue("Correct! You solved the riddle!"))
			team.Score = team.Score + 5
			saveTeamToFirebase(*team)
		} else {
			wrongGuesses++
			fmt.Println(red("Incorrect guess!"))
			fmt.Println(blue("The correct answer was: ", riddle.Answer))
			drawHangman(wrongGuesses)
		}

		fmt.Printf("Team %s Score: %d\n", team.Name, team.Score)
	}

	saveTeamToFirebase(*team)
}

func userInterface() {
	reader := bufio.NewReader(os.Stdin)
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	displaysolarisLogo()

	// Fetch approved teams from Firebase
	approvedTeams, err := getApprovedTeamsFromFirebase()
	if err != nil {
		log.Fatalf("Error fetching approved teams: %v\n", err)
	}

	var teamName string
	var team *Team
	teamEntered := false
	passwordVerified := false
	adminpasswordVerified := false

	for {
		if !adminpasswordVerified {
			fmt.Print(green("Enter admin the password to start the game: "))
			passwordEntered, _ := reader.ReadString('\n')
			passwordEntered = strings.TrimSpace(passwordEntered)

			correctPassword, err := getPasswordFromFirebase()
			if err != nil {
				fmt.Printf("Error retrieving password: %v\n", err)
				continue
			}

			if passwordEntered != correctPassword {
				fmt.Println(red("Incorrect password. Please try again."))
				continue
			}

			adminpasswordVerified = true // Mark the password as verified
		}
		if !teamEntered {
			// Prompt the user to enter their team name
			fmt.Print(green("Enter your team name: "))
			teamName, _ = reader.ReadString('\n')
			teamName = strings.TrimSpace(strings.ToLower(teamName)) // Remove leading/trailing spaces; no lowercase conversion

			// Validate the team name exactly as entered (case-sensitive)
			if !validateTeam(approvedTeams, teamName) {
				fmt.Println(red("Your team is not on the approved list. Contact admin for access."))
				continue
			}

			// Fetch the team from Firebase (no lowercase conversion)
			existingTeam, err := getTeamFromFirebase(teamName)
			if err == nil {
				// Team exists, retrieve it from Firebase
				team = &existingTeam

				// Check if attempts are increasing
				if team.Attempts > 0 {
					// If attempts are increasing, reset score to zero
					team.Score = 0
				}
				team.Attempts++ // Increment attempts
				fmt.Println(blue("Existing team found. Attempts incremented and score reset to 0."))
			} else {
				// Team doesn't exist, create a new team
				team = &Team{Name: teamName, Score: 0, Attempts: 1}
				fmt.Println(blue("Team not found. Creating a new team..."))
				createPasswordForNewTeam(team, reader)
				passwordVerified = true
			}

			teamEntered = true
		}

		if teamEntered && !passwordVerified {
			if team.Password != "" {
				// Validate existing password
				if validatePassword(team, reader) {
					passwordVerified = true
				} else {
					fmt.Println(red("Incorrect password. Please try again."))
					continue
				}
			}
		}

		if passwordVerified {
			// Game logic starts here
			fmt.Print(green("Type 'run' to start the game or 'close' to exit: "))
			command, _ := reader.ReadString('\n')
			command = strings.TrimSpace(strings.ToLower(command))

			if command == "run" {
				startScoreUpdater(team)

				// Get game duration from Firebase
				gameDuration, err := getGameDurationFromFirebase()
				if err != nil {
					log.Printf("Error getting game duration: %v. Using default of 5 minutes.\n", err)
					gameDuration = 5 * time.Minute
				}

				// Display the total time allotted
				minutes := int(gameDuration.Minutes())
				seconds := int(gameDuration.Seconds()) % 60
				fmt.Printf("\n%s You will have %s to solve all riddles.\n\n", yellow("Time Allotted:"), yellow(fmt.Sprintf("%dmin %dsec", minutes, seconds)))

				// Start the timer with the duration from Firebase
				startTimer(gameDuration)

				// Fetch riddles from Firebase or hardcoded ones
				riddlesSubset, err := randomRiddles(15)
				if err != nil {
					log.Fatalf("Error fetching riddles: %v\n", err)
				}

				// Run riddles with a timer
				runRiddles(team, riddlesSubset, reader)

				for {
					fmt.Print(green("Type 'close' to exit: "))
					command, _ := reader.ReadString('\n')
					command = strings.TrimSpace(strings.ToLower(command))

					if command == "close" {
						fmt.Println(blue("Exiting the game..."))
						return
					} else {
						fmt.Println(red("Invalid command. Please type 'close'."))
					}
				}
			} else if command == "close" {
				fmt.Println(blue("Exiting..."))
				break
			} else {
				fmt.Println(red("Invalid command"))
			}
		}
	}
}

func main() {
	initFirebase()
	userInterface()
}
