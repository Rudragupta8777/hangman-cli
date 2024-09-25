const firebaseConfig = {
    apiKey: "AIzaSyBqGpI2FfrDQ7X3DYcoyKNsDuav9EUQ7rg",
    authDomain: "hangman-cli.firebaseapp.com",
    databaseURL: "https://hangman-cli-default-rtdb.firebaseio.com",
    projectId: "hangman-cli",
    storageBucket: "hangman-cli.appspot.com",
    messagingSenderId: "613562680386",
    appId: "1:613562680386:web:0361d7d708ad0f97a559db"
};

try {
    firebase.initializeApp(firebaseConfig);
    const db = firebase.firestore();

    function updateScores() {
        db.collection("teams").onSnapshot((snapshot) => {
            let teams = [];
            snapshot.forEach((doc) => {
                teams.push({ id: doc.id, ...doc.data() });
            });

            if (teams.length === 0) {
                document.getElementById('errorMessage').textContent = "No teams found in the database.";
                return;
            }

            // Sort teams by score in descending order (highest score first)
            teams.sort((a, b) => (b.score || 0) - (a.score || 0));

            const scoreBoard = document.getElementById('scoreBoard');
            
            // Clear previous content
            scoreBoard.innerHTML = '';
            
            // Add the header
            const headerRow = document.createElement('div');
            headerRow.className = 'scoreboard-header';
            headerRow.innerHTML = `
                <span>Place</span>
                <span>Team Name</span>
                <span>Score</span>
            `;
            scoreBoard.appendChild(headerRow);

            // Update the top 3 teams with 3D prize effects
            if (teams[0]) {
                document.getElementById('first-team').textContent = teams[0].name || 'Unknown';
                document.getElementById('first-score').textContent = teams[0].score != null ? `Score: ${teams[0].score}` : 'N/A';
            }
            if (teams[1]) {
                document.getElementById('second-team').textContent = teams[1].name || 'Unknown';
                document.getElementById('second-score').textContent = teams[1].score != null ? `Score: ${teams[1].score}` : 'N/A';
            }
            if (teams[2]) {
                document.getElementById('third-team').textContent = teams[2].name || 'Unknown';
                document.getElementById('third-score').textContent = teams[2].score != null ? `Score: ${teams[2].score}` : 'N/A';
            }

            // Display teams in the leaderboard, starting from the 4th place
            teams.slice(3).forEach((team, index) => {
                const teamElement = document.createElement('div');
                teamElement.className = 'team';
                teamElement.innerHTML = `
                    <span>${index + 4}</span>
                    <span class="team-name">${team.name || 'Unknown'}</span>
                    <span class="team-score">${team.score != null ? team.score : 'N/A'}</span>
                `;
                scoreBoard.appendChild(teamElement);
            });

            console.log("Scores updated successfully");
        }, (error) => {
            console.error("Error fetching scores:", error);
            document.getElementById('errorMessage').textContent = "Error fetching scores. Please check the console for details.";
        });
    }

    updateScores();
    console.log("Script loaded and executed");
} catch (error) {
    console.error("Error initializing Firebase:", error);
    document.getElementById('errorMessage').textContent = "Error initializing Firebase. Please check the console for details.";
}