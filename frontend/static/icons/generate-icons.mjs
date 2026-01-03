/**
 * Icon Generator Script for AddressIQ
 *
 * Generates PNG icons from SVG sources for PWA/home screen support.
 * Run with: node generate-icons.mjs
 *
 * Prerequisites: npm install sharp
 */

import sharp from 'sharp';
import { readFileSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Brand colours
const GRADIENT_START = '#4A90D9';
const GRADIENT_END = '#357ABD';

/**
 * Creates an SVG icon with the AddressIQ branding
 * @param {number} size - Icon dimensions
 * @param {boolean} maskable - Whether to use maskable safe zone
 * @param {boolean} simplified - Whether to use simplified version (favicon)
 * @returns {string} SVG content
 */
function createIconSVG(size, maskable = false, simplified = false) {
    const cornerRadius = maskable ? 0 : Math.round(size * 0.176);
    const centre = size / 2;

    if (simplified) {
        // Simplified favicon - just centred "A"
        return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <defs>
    <linearGradient id="bgGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:${GRADIENT_START}"/>
      <stop offset="100%" style="stop-color:${GRADIENT_END}"/>
    </linearGradient>
  </defs>
  <rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="url(#bgGrad)"/>
  <text x="${centre}" y="${size * 0.82}"
        font-family="Inter, Arial, sans-serif"
        font-size="${size * 0.85}"
        font-weight="800"
        text-anchor="middle"
        fill="#FFFFFF">A</text>
</svg>`;
    }

    // Full icon with centred A and IQ between the legs
    const scale = maskable ? 0.75 : 1;
    const offsetX = maskable ? size * 0.125 : 0;
    const offsetY = maskable ? size * 0.125 : 0;
    const contentCentre = maskable ? size * 0.5 : centre;

    const aFontSize = Math.round(size * 0.82 * scale);
    const iqFontSize = Math.round(size * 0.17 * scale);
    const aY = Math.round(offsetY + size * 0.78 * scale);
    const iqY = Math.round(offsetY + size * 0.77 * scale);

    return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <defs>
    <linearGradient id="bgGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:${GRADIENT_START}"/>
      <stop offset="100%" style="stop-color:${GRADIENT_END}"/>
    </linearGradient>
  </defs>
  <rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="url(#bgGrad)"/>
  <text x="${contentCentre}" y="${aY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${aFontSize}"
        font-weight="800"
        text-anchor="middle"
        fill="#FFFFFF">A</text>
  <text x="${contentCentre}" y="${iqY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${iqFontSize}"
        font-weight="700"
        text-anchor="middle"
        fill="#FFFFFF"
        opacity="0.95">IQ</text>
</svg>`;
}

async function generateIcon(filename, size, maskable = false, simplified = false) {
    const svg = createIconSVG(size, maskable, simplified);
    const outputPath = join(__dirname, filename);

    await sharp(Buffer.from(svg))
        .resize(size, size)
        .png()
        .toFile(outputPath);

    console.log(`✓ Generated ${filename} (${size}×${size})`);
}

async function main() {
    console.log('Generating AddressIQ icons...\n');

    try {
        // Standard icons
        await generateIcon('icon-192.png', 192);
        await generateIcon('icon-512.png', 512);

        // Maskable icons (Android adaptive)
        await generateIcon('icon-maskable-192.png', 192, true);
        await generateIcon('icon-maskable-512.png', 512, true);

        // Apple touch icons
        await generateIcon('apple-touch-icon-152.png', 152);
        await generateIcon('apple-touch-icon-167.png', 167);
        await generateIcon('apple-touch-icon-180.png', 180);

        // Favicons
        await generateIcon('favicon-32.png', 32, false, true);
        await generateIcon('favicon-16.png', 16, false, true);

        console.log('\n✅ All icons generated successfully!');
        console.log('Icons are ready in: frontend/static/icons/');
    } catch (error) {
        console.error('Error generating icons:', error);
        process.exit(1);
    }
}

main();
