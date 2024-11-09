const createGame = async (player1, player2) => {
    let res = await fetch("https://connect4.avashist.com/match", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
            "Player1": player1,
            "Player2": player2,
            "StartPlayer": "", // You can set this to an actual value if needed
        }),
    });

    if (res.status != 200) {
        console.log("Some error occurred with the call");
        throw new Error("Failed to create the game");  // Explicitly throw an error for non-200 status codes
    }

    let data = await res.json();
    return data["match_id"];
};


const handleCreateGame = () => {
    let player1 = document.getElementById("playerAName").value;
    let player2 = document.getElementById("playerBName").value;
    
    createGame(player1, player2)
        .then((res) => {
            console.log(`Player 1 ${player1}, Player 2 ${player2}, and the match ID is ${res}`);
            window.location.href +=  `match/${res}`; 
        })
        .catch(() => {
            alert("Something went wrong");  // Use an arrow function here to defer execution
        });
};


const handleCreateLiveGame = () => {
    createGame("anonymous", "anonymous").then((res) => {
        console.log(`the match ID is ${res}`);
        window.location.href +=  `live/${res}`; 
    }).catch(() => {
        alert("Something went wrong")
    })
}
