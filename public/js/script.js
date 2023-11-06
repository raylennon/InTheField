const canvas = document.getElementById('myCanvas');
const ctx = canvas.getContext('2d');

ctx.canvas.width = window.innerWidth;
ctx.canvas.height = window.innerHeight;

scale = 12;
canvas.width = window.innerWidth/scale;
canvas.height = window.innerHeight/scale;

const width = canvas.width;
const height = canvas.height;

window.mouseX = window.innerWidth/2;
window.mouseY = window.innerHeight/2;

document.addEventListener('DOMContentLoaded', function() {
    document.addEventListener('mousemove', function(event) {
        // Store the cursor position in global variables
        window.mouseX = event.clientX;
        window.mouseY = event.clientY;
    });
});

window.imageData = ctx.createImageData(width, height).data;

window.updateFromBuffer = function (buffer) {

    const imageData = new ImageData(new Uint8ClampedArray(buffer), width, height);

    ctx.putImageData(imageData, 0, 0);
}

// const go=null;
// const wasm=null;

const importedFunctions = {};

async function initializeWasm() {
    const go = new Go();
    const wasm = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
    go.run(wasm.instance);

    ctx.rect(20, 20, 150, 100);
    ctx.fillStyle = "red";
    ctx.fill();
    // updateGamestate();   
    importedFunctions.generateImage = generateImage;
    importedFunctions.updateGamestate = updateGamestate;
}

function gameLoop() {
    importedFunctions.updateGamestate(window.mouseX, window.mouseY)
    importedFunctions.generateImage()
}

initializeWasm().then(() => {
    setInterval(gameLoop, 100);
});

document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape' || event.key === 'Esc') {
        console.log("Pressed...")
        importedFunctions.updateGamestate(window.innerWidth/2,window.innerHeight/2,"Pause!")
    }
  });
