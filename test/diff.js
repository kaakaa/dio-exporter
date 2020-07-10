if (process.argv.length < 5) {
    console.log("ERROR: You must specified just 3 args ('node diff.js cmd img1 img2')");
    return
}

const fs = require('fs');
const PNG = require('pngjs').PNG;
const pixelmatch = require('pixelmatch');

// node diff.js [pixel|image] img1 img2
const cmd = process.argv[2];
const original = process.argv[3];
const comparison = process.argv[4];


if (cmd !== 'pixel' && cmd !== 'image') {
    console.log("ERROR: First argument must be 'pixel' or 'image'");
    return
}

if (cmd === 'pixel' && process.argv.length !== 5) {
    console.log("ERROR: You must specify just 3 args: 'node diff.js pixel img1 img2'");
    return
}
if (cmd === 'image' && process.argv.length !== 6) {
    console.log("ERROR: You must specify just 3 args: 'node diff.js image img1 img2 outputFile'");
    return
}

const img1 = PNG.sync.read(fs.readFileSync(original));
const img2 = PNG.sync.read(fs.readFileSync(comparison));
const {width, height} = img1;
const diff = new PNG({width, height});
const numDiffPixel = pixelmatch(img1.data, img2.data, diff.data, width, height, {threshold: 0.1});


if (cmd === 'pixel') {
    console.log(numDiffPixel);
} else if (cmd === 'image') {
    const output = process.argv[5];
    fs.writeFileSync(output, PNG.sync.write(diff));
}

