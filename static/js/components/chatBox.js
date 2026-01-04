class ChatBox extends HTMLElement {
    connectedCallback() {
        this.innerHTML = `
            <style>
                #chatBox {
                    background-color: #E16A54;
                    border-radius: 12px;
                    padding: 20px;
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
                    max-width: 400px;
                    margin: 20px auto;
                }
                @media (max-width: 768px) {
                    #chatBox {
                        position: fixed;
                        bottom: 0;
                        left: 0;
                        right: 0;
                        max-width: 100%;
                        margin: 0;
                        border-radius: 16px 16px 0 0;
                        max-height: 70vh;
                        display: flex;
                        flex-direction: column;
                        z-index: 1000;
                        transform: translateY(calc(100% - 60px));
                        transition: transform 0.3s ease;
                        padding-bottom: 80px;
                    }
                    #chatBox.expanded {
                        padding-bottom: 0;
                    }
                    #chatBox.expanded {
                        transform: translateY(0);
                    }
                    #chatBox.expanded ~ #chatToggleButton,
                    body:has(#chatBox.expanded) #chatToggleButton {
                        opacity: 0 !important;
                        pointer-events: none !important;
                        transform: scale(0.8) !important;
                    }
                    #chatToggleButton.hidden {
                        display: none;
                    }
                    #chatBox h1 {
                        font-size: 18px;
                        padding-bottom: 8px;
                        margin-bottom: 12px;
                        cursor: pointer;
                        user-select: none;
                        position: relative;
                        padding-right: 30px;
                    }
                    #chatBox h1::after {
                        content: '✕';
                        position: absolute;
                        right: 0;
                        top: 50%;
                        transform: translateY(-50%);
                        font-size: 20px;
                        opacity: 0.8;
                        transition: opacity 0.2s ease;
                    }
                    #chatBox h1:hover::after {
                        opacity: 1;
                    }
                    #chatBox:not(.expanded) h1::after {
                        display: none;
                    }
                    #chatMessages {
                        height: calc(70vh - 160px);
                        flex: 1;
                        min-height: 200px;
                    }
                    #chatToggleButton {
                        display: block;
                    }
                }
                @media (min-width: 769px) {
                    #chatToggleButton {
                        display: none;
                    }
                }
                #chatToggleButton {
                    position: fixed;
                    bottom: 90px;
                    right: 20px;
                    width: 56px;
                    height: 56px;
                    background-color: #7C444F;
                    color: white;
                    border: none;
                    border-radius: 50%;
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
                    cursor: pointer;
                    z-index: 1001;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 24px;
                    transition: all 0.3s ease;
                }
                @media (max-width: 768px) {
                    #chatBox.expanded ~ #chatToggleButton {
                        bottom: 20px;
                    }
                }
                #chatToggleButton:hover {
                    transform: scale(1.1);
                    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
                }
                #chatToggleButton:active {
                    transform: scale(0.95);
                }
                #chatBox h1 {
                    color: #7C444F;
                    font-size: 24px;
                    font-weight: 700;
                    margin: 0 0 16px 0;
                    text-align: center;
                    border-bottom: 2px solid rgba(124, 68, 79, 0.3);
                    padding-bottom: 12px;
                }
                #chatMessages {
                    height: 300px;
                    width: 100%;
                    overflow-y: auto;
                    background-color: rgba(255, 255, 255, 0.95);
                    border-radius: 8px;
                    padding: 12px;
                    margin-bottom: 12px;
                    box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.1);
                }
                #chatMessages::-webkit-scrollbar {
                    width: 8px;
                }
                #chatMessages::-webkit-scrollbar-track {
                    background: rgba(0, 0, 0, 0.05);
                    border-radius: 4px;
                }
                #chatMessages::-webkit-scrollbar-thumb {
                    background: rgba(124, 68, 79, 0.5);
                    border-radius: 4px;
                }
                #chatMessages::-webkit-scrollbar-thumb:hover {
                    background: rgba(124, 68, 79, 0.7);
                }
                .message {
                    background: linear-gradient(135deg, rgba(124, 68, 79, 0.15) 0%, rgba(124, 68, 79, 0.08) 100%);
                    border-left: 3px solid #7C444F;
                    border-radius: 8px;
                    padding: 10px 14px;
                    margin-bottom: 10px;
                    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
                    transition: all 0.2s ease;
                    animation: slideIn 0.3s ease;
                }
                .message:hover {
                    transform: translateX(4px);
                    box-shadow: 0 3px 6px rgba(0, 0, 0, 0.15);
                }
                .message:last-child {
                    margin-bottom: 0;
                }
                .message-player {
                    font-weight: 700;
                    color: #7C444F;
                    font-size: 14px;
                    margin-right: 8px;
                }
                .message-text {
                    color: #333;
                    font-size: 14px;
                    line-height: 1.4;
                }
                @keyframes slideIn {
                    from {
                        opacity: 0;
                        transform: translateY(-10px);
                    }
                    to {
                        opacity: 1;
                        transform: translateY(0);
                    }
                }
                #chatInputContainer {
                    display: flex;
                    flex-direction: row;
                    align-items: center;
                    gap: 8px;
                }
                #chatInput {
                    flex: 1;
                    padding: 12px 16px;
                    border: 2px solid rgba(124, 68, 79, 0.3);
                    border-radius: 8px;
                    background-color: rgba(255, 255, 255, 0.95);
                    font-size: 14px;
                    transition: all 0.2s ease;
                }
                #chatInput:focus {
                    outline: none;
                    border-color: #7C444F;
                    box-shadow: 0 0 0 3px rgba(124, 68, 79, 0.1);
                }
                #chatInput::placeholder {
                    color: rgba(124, 68, 79, 0.5);
                }
                #sendButton {
                    padding: 12px 24px;
                    background-color: #7C444F;
                    color: white;
                    border: none;
                    border-radius: 8px;
                    font-weight: 600;
                    font-size: 14px;
                    cursor: pointer;
                    transition: all 0.2s ease;
                    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
                }
                #sendButton:hover {
                    background-color: rgba(124, 68, 79, 0.9);
                    transform: translateY(-2px);
                    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
                }
                #sendButton:active {
                    transform: translateY(0);
                    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
                }
                #sendButton:disabled {
                    opacity: 0.5;
                    cursor: not-allowed;
                    transform: none;
                }
            </style>
            <button id="chatToggleButton" aria-label="Toggle Chat">💬</button>
            <div id="chatBox">
                <h1 id="chatHeader">Chat</h1>
                <div id="chatMessages">
                </div>
                <div id="chatInputContainer">
                    <input type="text" id="chatInput" placeholder="Type your message here">
                    <button id="sendButton">Send</button>
                </div>
            </div>
        `
        const sendMessage = () => {
            const input = this.querySelector('#chatInput');
            const message = input.value.trim();
            if (message !== "") {
                // Dispatch custom event to parent
                this.dispatchEvent(new CustomEvent('chat-message', {
                    detail: { message: message },
                    bubbles: true
                }));
                input.value = ''; // Clear input
            }
        };

        this.querySelector('#sendButton').addEventListener('click', sendMessage);
        
        this.querySelector('#chatInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });

        // Mobile toggle functionality
        const chatBox = this.querySelector('#chatBox');
        const toggleButton = this.querySelector('#chatToggleButton');
        const chatHeader = this.querySelector('#chatHeader');
        
        const toggleChat = () => {
            chatBox.classList.toggle('expanded');
            // Update toggle button visibility
            if (toggleButton && window.innerWidth <= 768) {
                if (chatBox.classList.contains('expanded')) {
                    toggleButton.style.opacity = '0';
                    toggleButton.style.pointerEvents = 'none';
                    toggleButton.style.transform = 'scale(0.8)';
                } else {
                    toggleButton.style.opacity = '1';
                    toggleButton.style.pointerEvents = 'auto';
                    toggleButton.style.transform = 'scale(1)';
                }
            }
        };
        
        if (toggleButton) {
            toggleButton.addEventListener('click', (e) => {
                e.stopPropagation();
                toggleChat();
            });
        }
        
        // Always allow header click on mobile to close chat
        if (chatHeader) {
            chatHeader.addEventListener('click', (e) => {
                if (window.innerWidth <= 768) {
                    e.stopPropagation();
                    toggleChat();
                }
            });
        }
    }

    addMessage = (message) => {
        const chatMessages = document.getElementById('chatMessages');
        const messageElement = document.createElement('div');
        messageElement.classList.add('message');
        messageElement.innerHTML = `
            <span class="message-player">${message.name || message.player}:</span>
            <span class="message-text">${message.message}</span>
        `;
        chatMessages.appendChild(messageElement);
        // Auto-scroll to bottom
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
}

customElements.define('chat-box', ChatBox);