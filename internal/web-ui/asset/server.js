/**
 * @param {string} url
 */
export const connect = (url) => {
  const socket = new WebSocket(url);

  /** @param {string} string to write to WS connection */
  const write = (string) => {
    socket.send(string);
  };
  const writeMessage = (type, data) => {
    const m = {
      v: 1,
      type,
      data,
    }
    socket.send(JSON.stringify(m))
  }
  socket.addEventListener('open', function (event) {
    write('Hello Server!');
  });
  socket.addEventListener('close', function () {
    console.log('closed');
  });
  const messageListeners = [];
  const addMessageListener = (listner) => {
    messageListeners.push(listner);
  };
  // Listen for messages
  socket.addEventListener('message', function (event) {
    console.log('Message from server ', event.data);
    const message = JSON.parse(event.data);
    messageListeners.forEach((x) => x(message));
  });
  return {
    write,
    writeMessage,
    addMessageListener,
  };
};

export const connectionHealthIndicator = (() => {
  const symbol = {
    heart: '❤️',
    grave: '🪦',
    blue: '💙',
    purple: '💜',
    black: '🖤',
  };
  const decayTimeout = {
    purple: undefined,
    blue: undefined,
    black: undefined,
    grave: undefined,
  };
  function getElement() {
    return document.getElementById('heart');
  }
  function timeReceived() {
    var heart = getElement();
    if (!heart) return;
    setAlive();
    window.requestAnimationFrame(function () {
      heart.style.fontSize = '110%';
      heart.style.transition = 'font-size 150ms';
      setTimeout(function () {
        heart.style.transition = 'font-size 400ms';
        heart.style.fontSize = '100%';
      }, 400);
      clearDecayTimeouts();
      startDecayTimeouts();
    });
  }
  function startDecayTimeouts() {
    decayTimeout.purple = setTimeout(startDecay(symbol.purple), 1100);
    decayTimeout.blue = setTimeout(startDecay(symbol.blue), 2500);
    decayTimeout.black = setTimeout(startDecay(symbol.black), 4000);
    decayTimeout.grave = setTimeout(startDecay(symbol.grave), 5500);
  }
  function startDecay(symbol) {
    if (!getElement()) return;
    return function () {
      getElement().innerText = symbol;
    };
  }
  function clearDecayTimeouts() {
    clearTimeout(decayTimeout.purple);
    clearTimeout(decayTimeout.blue);
    clearTimeout(decayTimeout.black);
    clearTimeout(decayTimeout.grave);
  }
  function setAlive() {
    if (!getElement()) return;
    clearDecayTimeouts();
    getElement().innerText = symbol.heart;
  }
  return (message) => {
    if (message.type === 'time') {
      timeReceived();
    }
  };
})();
