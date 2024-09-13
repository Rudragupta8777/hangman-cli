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
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type Riddle struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

var riddles = []Riddle{
	{"I speak without a mouth and hear without ears. I have no body, but I come alive with wind.", "echo"},
	{"The more of this there is, the less you see.", "darkness"},
	{"What has keys but can't open locks?", "piano"},
	{"The more you take, the more you leave behind.", "footsteps"},
	{"What has to be broken before you can use it?", "egg"},
	{"I'm tall when I'm young, and I'm short when I'm old. What am I?", "candle"},
	{"What month of the year has 28 days?", "all"},
	{"What is full of holes but still holds water?", "sponge"},
	{"What question can you never answer yes to?", "are you asleep"},
	{"What is always in front of you but can't be seen?", "future"},
	{"There's a one-story house in which everything is yellow. Yellow walls, yellow doors, yellow furniture. What color are the stairs?", "no stairs"},
	{"What can you break, even if you never pick it up or touch it?", "promise"},
	{"I'm light as a feather, yet the strongest person can't hold me for five minutes. What am I?", "breath"},
	{"I'm found in socks, scarves and mittens; and often in the paws of playful kittens. What am I?", "yarn"},
	{"Where does today come before yesterday?", "dictionary"},
	{"What invention lets you look right through a wall?", "window"},
	{"If you have me, you want to share me. If you share me, you don't have me. What am I?", "secret"},
	{"What goes up but never comes down?", "age"},
	{"The more you take, the more you leave behind. What am I?", "footsteps"},
	{"What can travel around the world while staying in one corner?", "stamp"},
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

const firebaseCredentials = `` // copy paste the firebase credientials here

func initFirebase() {
	opt := option.WithCredentialsJSON([]byte(firebaseCredentials))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing app: %v\n", err)
	}
	firebaseApp = app
}

func saveTeamScoreToFirebase(team Team) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	// Update the score and name inside the document for the given team
	_, err = client.Collection("teams").Doc(team.Name).Set(ctx, map[string]interface{}{
		"score": team.Score,
		"name":  team.Name,
	}, firestore.MergeAll) // Use firestore.MergeAll to merge the fields

	if err != nil {
		log.Fatalf("Error updating score in Firebase: %v\n", err)
	}
}

func randomRiddles(num int) []Riddle {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(riddles), func(i, j int) {
		riddles[i], riddles[j] = riddles[j], riddles[i]
	})
	return riddles[:num]
}

func drawHangman(stage int) {
	fmt.Println(hangmanStages[stage])
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

func startScoreUpdater(team *Team) {
	go func() {
		for {
			time.Sleep(1 * time.Second) // Update score every 1 second
			saveTeamScoreToFirebase(*team)
		}
	}()
}

func displayLogo() {
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

func userInterface() {
	reader := bufio.NewReader(os.Stdin)
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	displayLogo()

	var teamName string
	var team *Team
	teamEntered := false

	for {
		if !teamEntered {
			fmt.Print(green("Enter your team name: "))
			teamName, _ = reader.ReadString('\n')
			teamName = strings.TrimSpace(teamName)
			team = &Team{Name: teamName, Score: 0}
			teamEntered = true
		}

		fmt.Print(green("Enter the password to start the game: "))
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

		fmt.Print(green("Type 'run' to start the game or 'close' to exit: "))
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(strings.ToLower(command))

		if command == "run" {
			// Start a score updater goroutine to update the score every second
			startScoreUpdater(team)

			riddlesSubset := randomRiddles(10)
			wrongGuesses := 0
			for i, riddle := range riddlesSubset {
				if wrongGuesses >= len(hangmanStages)-1 {
					fmt.Println(red("You've been hanged!"))
					drawHangman(wrongGuesses)

					for {
						fmt.Print(green("Type 'close' to exit or 'retry' to play again: "))
						exitCommand, _ := reader.ReadString('\n')
						exitCommand = strings.TrimSpace(strings.ToLower(exitCommand))

						if exitCommand == "close" {
							fmt.Println("Exiting the game...")
							return
						} else if exitCommand == "retry" {
							fmt.Println("Starting a new game...")
							break
						} else {
							fmt.Println(red("Invalid command. Please type 'close' or 'retry'."))
						}
					}

					break
				}

				fmt.Printf("\n%s %s\n", green("Question "+fmt.Sprintf("%d:", i+1)), riddle.Question)
				fmt.Print(green("Enter your guess [whole word]: "))
				guess, _ := reader.ReadString('\n')
				guess = strings.TrimSpace(strings.ToLower(guess))

				if guess == strings.ToLower(riddle.Answer) {
					fmt.Println(blue("Correct! You solved the riddle!"))
					team.Score++
					saveTeamScoreToFirebase(*team) // Update Firebase immediately after correct guess
				} else {
					wrongGuesses++
					fmt.Println(red("Incorrect guess!"))
					fmt.Println(green("The correct answer was: ", riddle.Answer))
					drawHangman(wrongGuesses)
				}

				// Print the score only once, after each guess
				fmt.Printf("Team %s Score: %d\n", team.Name, team.Score)
			}

			saveTeamScoreToFirebase(*team)

		} else if command == "close" {
			fmt.Println("Exiting...")
			break
		} else {
			fmt.Println(red("Invalid command"))
		}
	}
}

func main() {
	initFirebase()  // Initialize Firebase
	userInterface() // Run user interface
}
