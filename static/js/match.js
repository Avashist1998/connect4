var currPlayer = document.getElementById("player1").innerHTML;
var currSlot = document.getElementById("currSlot").innerHTML;

const addColumnClickListeners = () => {
    const columns = document.querySelectorAll('.col');
    columns.forEach(col => {
        col.addEventListener("click", () => {
            const col_number = Number(col.id.replace("col-", ""))
            currPlayer = document.getElementById("currPlayer").innerHTML; 
            console.log(`Player: ${currPlayer}, Slot: ${currSlot} Column Clicked: ${col_number}`)
            makeMove(currPlayer, currSlot, col_number)
        })
    })
}


const makeMove = async (player, slot, move) => {
    let res = await fetch(window.location.href, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
            "Player": player,
            "Slot": slot,
            "Move": Number(move),
        }),
    });

    if (res.status != 200) {
        console.log("Some error occurred with the call");
        throw new Error("Failed to create the game");  
    }

    let data = await res.json();
    currPlayer = document.getElementById("currPlayer")
    currSlot = data.currSlot;
    // currPlayer.innerHTML = data.currPlayer
    const boardElement = document.getElementById('gameBoard');
    boardElement.updateGrid(data.board);
    if (data.message == "Game Over") {
        document.querySelector(".main").classList += " disabled"
        document.getElementById("gameOverModal").style.display = "block"
        if (data.winner === "") {
            document.getElementById("gameOverHeading").innerHTML = "Draw"
        } else {
            document.getElementById("gameOverHeading").innerHTML = `Winner: ${data.winner}`
        }
    }

}

const handleClickHome = () => {
    window.location.href =  config.httpURL
}

const handleClickRematch = async () => {
    let res = await fetch(window.location.href, {
        method: "DELETE",
    });
    if (res.status != 200) {
        console.log("Some error occurred with the call");
        throw new Error("Failed to reset the game");
    }
    window.location.reload();
} 

document.addEventListener("DOMContentLoaded", () => {
    addColumnClickListeners();
    document.getElementById("homeButton").addEventListener("click", handleClickHome)
    document.getElementById("rematchButton").addEventListener("click", handleClickRematch)
});

