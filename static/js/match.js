

var currPlayer = document.getElementById("startPlayer").innerHTML.replace("Turn: ", "");
console.log(`the current player is ${currPlayer}`)

const addColumnClickListeners = () => {
    const columns = document.querySelectorAll('.col');
    columns.forEach(col => {
        col.addEventListener("click", () => {
            const col_number = col.id.replace("col-", "") 
            console.log(`Column Clicked: ${col_number}`)
            makeMove(currPlayer, col_number).then(() => {
                console.log("This is a person")
                window.location.reload();
            })
        })
    })
}

const makeMove = async (player, move) => {
    let res = await fetch(window.location.href, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
            "Player": player,
            "Move": Number(move),
        }),
    });

    if (res.status != 200) {
        console.log("Some error occurred with the call");
        throw new Error("Failed to create the game");
    }

    let data = await res.json();
    currPlayer = data["CurrPlayer"];
}


const resetGame = async () => {
    let res = await fetch(window.location.href, {
        method: "DELETE",
    });
    if (res.status != 200) {
        console.log("Some error occurred with the call");
        throw new Error("Failed to reset the game");
    }
    console.log("I made it here and I was called")
    window.location.reload();
} 

const addRestartEventHandler = () => {
    const reset = document.querySelector(".resetGame");
    if (reset !== null) {
        reset.addEventListener("click", () => {
            resetGame()
        })
    }
    const newGame = document.querySelector(".newGame");

    if (newGame !== null) {
        newGame.addEventListener("click", () => {
            window.location.href = "http://localhost:9080"
        })
    }
}


addColumnClickListeners()
addRestartEventHandler()