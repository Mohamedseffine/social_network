let socket = null;
const ports = new Set();
let wsUrl = '';

const connect = () => {
    if (socket) return;

    // The worker doesn't have access to window.location, so the URL must be sent from the main thread.
    // For now, we will hardcode it, but a robust solution would receive it via a message.
    // This needs to be configured based on the actual deployment environment.
    // Assuming the base URL is localhost:8080 for the backend.
    if (!wsUrl) {
        console.error("WebSocket URL not provided to worker.");
        return;
    }

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log("Shared Worker: WebSocket connection established.");
        ports.forEach(port => port.postMessage({ type: 'WS_OPEN' }));
    };

    socket.onmessage = (event) => {
        // Broadcast incoming messages to all connected tabs
        ports.forEach(port => {
            port.postMessage(JSON.parse(event.data));
        });
    };

    socket.onerror = (error) => {
        console.error("Shared Worker: WebSocket error:", error);
        ports.forEach(port => port.postMessage({ type: 'WS_ERROR', payload: 'WebSocket error' }));
    };

    socket.onclose = () => {
        console.log("Shared Worker: WebSocket connection closed.");
        socket = null; // Clear the socket
        ports.forEach(port => port.postMessage({ type: 'WS_CLOSE' }));
        // Optional: attempt to reconnect after a delay
        // setTimeout(connect, 5000);
    };
};

self.onconnect = (e) => {
    const port = e.ports[0];
    ports.add(port);
    console.log(`Shared Worker: New tab connected. Total ports: ${ports.size}`);

    port.onmessage = (event) => {
        const message = event.data;

        if (message.type === 'INIT_WS') {
            // Initialize with the WebSocket URL from the main thread
            wsUrl = message.payload;
            if (!socket || socket.readyState === WebSocket.CLOSED) {
                connect();
            }
        } else if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(message));
        } else {
            console.warn("Shared Worker: WebSocket is not open. Message not sent.", message);
            // Optionally, queue the message and send it upon reconnection.
        }
    };

    port.start();

    // Notify the newly connected tab about the current WebSocket state
    if (socket && socket.readyState === WebSocket.OPEN) {
        port.postMessage({ type: 'WS_OPEN' });
    }
};
