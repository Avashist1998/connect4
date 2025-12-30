class ChatBox extends HTMLElement {
    connectedCallback() {
        this.innerHTML = `
            <style>
                #chatInputContainer {
                    display: flex;
                    flex-direction: row;
                    align-items: center;
                    justify-content: center;
                }
                #chatInput {
                    width: 100%;
                    padding: 10px;
                    border: 1px solid #ccc;
                    border-radius: 5px;
                }
                #sendButton {
                    padding: 10px;
                }
                #chatMessages {
                    height: 300px;
                    width: 100%;
                    overflow-y: auto;
                    border: 1px solid #ccc;
                    border-radius: 5px;
                    padding: 10px;
                }
            </style>
            <div id="chatBox">
                <h1>Chat</h1>
                <div id="chatMessages">
                </div>
                <div id="chatInputContainer">
                    <input type="text" id="chatInput" placeholder="Type your message here">
                    <button id="sendButton">Send</button>
                </div>
            </div>
        `
        this.querySelector('#sendButton').addEventListener('click', () => {
            const input = this.querySelector('#chatInput');
            const message = input.value;
            if (message !== "") {
                // Dispatch custom event to parent
                this.dispatchEvent(new CustomEvent('chat-message', {
                    detail: { message: message },
                    bubbles: true
                }));
                input.value = ''; // Clear input
            }
        });
    }

    addMessage = (message) => {
        const chatMessages = document.getElementById('chatMessages');
        const messageElement = document.createElement('div');
        messageElement.classList.add('message');
        messageElement.innerHTML = message.player + ': ' + message.message;
        chatMessages.appendChild(messageElement);
    }



}

customElements.define('chat-box', ChatBox);