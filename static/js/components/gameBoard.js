class Board extends HTMLElement {
    connectedCallback() {
        this.innerHTML = `
            <style>
                .board {
                    display: flex;
                    flex-direction: row;
                    align-items: center;
                    justify-content: center;
                }
                
                .col {
                    display: flex;
                    flex-direction: column;
                    display: inline-block;
                }
                .col:hover {
                    /* background-color: gray; */
                    opacity: 0.5;
                }
                
                .cell {
                    width: 60px;
                    height: 60px;
                    position: relative;
                    background-color: blue;
                    display: flex;
                    justify-content: center;
                    align-items: center;
                }
                
                .cell::before {
                    content: "";
                    width: 50px;
                    height: 50px;
                    border-radius: 50%;
                    border: 2px solid black;
                }
                
                .player1::before {
                    background-color: red;   
                }
                
                .player2::before {
                    background-color: yellow;
                }
                
                .empty::before {
                    background-color: white;
                }
            
            </style>
            <div class="board">
                <div class="col" id="col-0">
                    <div class="cell empty" id="cell-35">
                    </div>
                    <div class="cell empty" id="cell-28">
                    </div>
                    <div class="cell empty" id="cell-21">
                    </div>
                    <div class="cell empty" id="cell-14">
                    </div>
                    <div class="cell empty" id="cell-7">
                    </div>
                    <div class="cell empty" id="cell-0">
                    </div>
                </div>
                <div class="col" id="col-1">
                    <div class="cell empty" id="cell-36">
                    </div>
                    <div class="cell empty" id="cell-29">
                    </div>
                    <div class="cell empty" id="cell-22">
                    </div>
                    <div class="cell empty" id="cell-15">
                    </div>
                    <div class="cell empty" id="cell-8">
                    </div>
                    <div class="cell empty" id="cell-1">
                    </div>
                </div>
                <div class="col" id="col-2">
                    <div class="cell empty" id="cell-37">
                    </div>
                    <div class="cell empty" id="cell-30">
                    </div>
                    <div class="cell empty" id="cell-23">
                    </div>
                    <div class="cell empty" id="cell-16">
                    </div>
                    <div class="cell empty" id="cell-9">
                    </div>
                    <div class="cell empty" id="cell-2">
                    </div>
                </div>
                <div class="col" id="col-3">
                    <div class="cell empty" id="cell-38">
                    </div>
                    <div class="cell empty" id="cell-31">
                    </div>
                    <div class="cell empty" id="cell-24">
                    </div>
                    <div class="cell empty" id="cell-17">
                    </div>
                    <div class="cell empty" id="cell-10">
                    </div>
                    <div class="cell empty" id="cell-3">
                    </div>
                </div>
                <div class="col" id="col-4">
                    <div class="cell empty" id="cell-39">
                    </div>
                    <div class="cell empty" id="cell-32">
                    </div>
                    <div class="cell empty" id="cell-25">
                    </div>
                    <div class="cell empty" id="cell-18">
                    </div>
                    <div class="cell empty" id="cell-11">
                    </div>
                    <div class="cell empty" id="cell-4">
                    </div>
                </div>
                <div class="col" id="col-5">
                    <div class="cell empty" id="cell-40">
                    </div>
                    <div class="cell empty" id="cell-33">
                    </div>
                    <div class="cell empty" id="cell-26">
                    </div>
                    <div class="cell empty" id="cell-19">
                    </div>
                    <div class="cell empty" id="cell-12">
                    </div>
                    <div class="cell empty" id="cell-5">
                    </div>
                </div>
                <div class="col" id="col-6">
                    <div class="cell empty" id="cell-41">
                    </div>
                    <div class="cell empty" id="cell-34">
                    </div>
                    <div class="cell empty" id="cell-27">
                    </div>
                    <div class="cell empty" id="cell-20">
                    </div>
                    <div class="cell empty" id="cell-13">
                    </div>
                    <div class="cell empty" id="cell-6">
                    </div>
                </div>
            </div>
        `
    }

    updateGrid = (board) => {
        for (let [index, val] of board.entries()) {
            let cell = document.getElementById(`cell-${index}`)
            if (val == 1) {
                cell.className = 'cell player1';
            } else if (val == -1) {
                cell.className = 'cell player2';
            } else {
                cell.className = 'cell empty';
            }
        }
    }
}

customElements.define('game-board', Board);