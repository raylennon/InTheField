const pressedKeys = { left: false, up: false, right: false, down: false };
const ids = { "a": "left", "w": "up", "d": "right", "s": "down", "Esc" : "pause", "e": "lspin"};

function handlePress(id, down) {
    if (pressedKeys[id] !== down) {
        pressedKeys[id] = down;
        let dir = ""; const baseUrl = dir;
        if (pressedKeys.left) { dir += "L"; }
        if (pressedKeys.right) { dir += "R"; }
        if (pressedKeys.up) { dir += "U"; }
        if (pressedKeys.down) { dir += "D"; }
        if (pressedKeys.lspin) { dir += "D"; }
        if (dir === baseUrl) { dir += "STOP"; }
        importedFunctions.updateGamestate(window.innerWidth / 2, window.innerHeight / 2, dir)
    }
}

document.addEventListener("keydown", (event) => {
    id = ids[event.key];
    if (id && !pressedKeys[id]) {
        handlePress(id, true);
    }
});

document.addEventListener("keyup", (event) => {
    id = ids[event.key];
    if (id && pressedKeys[id]) {
        handlePress(id, false);
    }
});