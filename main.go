package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/fatih/color"
	"github.com/google/uuid"

	firebase "firebase.google.com/go"
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

const firebaseCredentials = `{}` // copy paste the firebase credientials here

func initFirebase() {
	opt := option.WithCredentialsJSON([]byte(firebaseCredentials))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing app: %v\n", err)
	}
	firebaseApp = app
}

func GetMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || (iface.Flags&net.FlagLoopback != 0) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				return iface.HardwareAddr.String(), nil
			}
		}
	}

	return "", fmt.Errorf("unable to get MAC address")
}

func GenerateUUID() (string, error) {
	id := uuid.New()
	return id.String(), nil
}

func getMachineID() string {
	mac, err := GetMACAddress()
	if err != nil || mac == "" {
		id, err := GenerateUUID()
		if err != nil {
			log.Fatalf("Error generating UUID: %v\n", err)
		}
		return id
	}
	return mac
}

func saveTeamToFirebase(team Team) {
	ctx := context.Background()
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v\n", err)
	}
	defer client.Close()

	_, err = client.Collection("teams").Doc(team.Name).Set(ctx, map[string]interface{}{
		"score":     team.Score,
		"name":      team.Name,
		"machineID": team.MachineID,
		"attempts":  team.Attempts,
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

func userInterface() {
	reader := bufio.NewReader(os.Stdin)
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	displaysolarisLogo()

	var teamName string
	var team *Team
	teamEntered := false
	passwordVerified := false // Variable to track if the password was verified

	for {
		// Only ask for the password once, at the start
		if !passwordVerified {
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

			passwordVerified = true // Mark the password as verified
		}
		if !teamEntered {
			fmt.Print(green("Enter your team name: "))
			teamName, _ = reader.ReadString('\n')
			teamName = strings.TrimSpace(teamName)

			existingTeam, err := getTeamFromFirebase(teamName)
			if err == nil {
				// Team exists, check machine ID
				currentMachineID := getMachineID()
				if existingTeam.MachineID != currentMachineID {
					fmt.Println(red("Error: This team name is associated with a different machine. Please use a different team name or play on the original machine."))
					continue
				}
				team = &existingTeam
				team.Attempts++ // Increment attempts
				team.Score = 0  // Reset score to zero for a new attempt
			} else {
				// New team
				team = &Team{Name: teamName, Score: 0, MachineID: getMachineID(), Attempts: 1}
			}
			saveTeamToFirebase(*team)
			teamEntered = true
		}

		// Main game loop
		fmt.Print(green("Type 'run' to start the game or 'close' to exit: "))
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(strings.ToLower(command))

		if command == "run" {
			startScoreUpdater(team)

			// fmt.Printf("Attempt #%d for team %s\n", team.Attempts, team.Name)

			riddlesSubset := randomRiddles(10)
			wrongGuesses := 0
			for i, riddle := range riddlesSubset {
				if wrongGuesses >= len(hangmanStages)-1 {
					fmt.Println(red("You've been hanged!"))
					// drawHangman(wrongGuesses)
					displaygameoverLogo()
					break // Exit the riddle loop, move to exit/retry options
				}

				fmt.Printf("\n%s %s\n", green("Question "+fmt.Sprintf("%d:", i+1)), riddle.Question)
				fmt.Print(green("Enter your guess [whole word]: "))
				guess, _ := reader.ReadString('\n')
				guess = strings.TrimSpace(strings.ToLower(guess))

				if guess == strings.ToLower(riddle.Answer) {
					fmt.Println(blue("Correct! You solved the riddle!"))
					team.Score = team.Score + 5
					saveTeamToFirebase(*team) // Update Firebase immediately after correct guess
				} else {
					wrongGuesses++
					fmt.Println(red("Incorrect guess!"))
					fmt.Println(blue("The correct answer was: ", riddle.Answer))
					drawHangman(wrongGuesses)
				}

				// Print the score only once, after each guess
				fmt.Printf("Team %s Score: %d\n", team.Name, team.Score)
			}

			saveTeamToFirebase(*team)

			// After the game ends
			// fmt.Printf("Game over! Final score for attempt #%d: %d\n", team.Attempts, team.Score)

			// Ask to play again or exit
			for {
				fmt.Print(green("Type 'close' to exit: "))
				command, _ := reader.ReadString('\n')
				command = strings.TrimSpace(strings.ToLower(command))

				if command == "close" {
					fmt.Println(blue("Exiting the game..."))
					return
				} else {
					fmt.Println(red("Invalid command. Please type 'play' or 'close'."))
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

func main() {
	initFirebase()  // Initialize Firebase
	userInterface() // Run user interface
}
