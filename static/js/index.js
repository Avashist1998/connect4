const BASE_URL = "https://connect4.avashist.com"

const createGame = async (type, player1, player2) => {
    let res = await fetch(BASE_URL, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
            "GameType": type,
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


const handleCreateLocalGame = () => {
    let player1 = document.getElementById("playerAName").value;
    let player2 = document.getElementById("playerBName").value;
    
    createGame("local", player1, player2)
        .then((matchID) => {
            console.log(`Player 1 ${player1}, Player 2 ${player2}, and the match ID is ${matchID}`);
            window.location.href +=  `${matchID}`; 
        })
        .catch(() => {
            alert("Something went wrong");  // Use an arrow function here to defer execution
        });
};


const handleCreateLiveGame = () => {
    createGame("live", "anonymous", "anonymous").then((matchID) => {
        console.log(`the match ID is ${matchID}`);
        window.location.href +=  `${matchID}`; 
    }).catch((e) => {
        alert("Something went wrong")
    })
}


const handleCreateBotGame = () => {
    createGame("bot", "anonymous", "anonymous").then((res) => {
        console.log(`the match ID is ${res}`);
        window.location.href +=  `live/${res}`; 
    }).catch((e) => {
        alert("Something went wrong")
    })
}