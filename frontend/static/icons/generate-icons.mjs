/**
 * Icon Generator Script for NethAddress
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
// Brand colours
const GRADIENT_START = '#2DD4BF';
const GRADIENT_END = '#0F766E';

/**
 * Creates an SVG icon with the NethAddress branding
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

/**
 * Creates an SVG icon for iOS 18 with transparent background
 * Works with Clear, Tinted, and Dark modes in Liquid Glass
 * Teal brand colour matching frontend theme, darkened for contrast
 */
function createAppleIconSVG(size) {
	const centre = size / 2;
	const aFontSize = Math.round(size * 0.82);
	const iqFontSize = Math.round(size * 0.17);
	const aY = Math.round(size * 0.79);
	const iqY = Math.round(size * 0.78);

	// Teal brand colour darkened for better contrast (#0d9488 -> #0A6B62)
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <text x="${centre}" y="${aY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${aFontSize}"
        font-weight="800"
        text-anchor="middle"
        fill="#0A6B62">A</text>
  <text x="${centre}" y="${iqY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${iqFontSize}"
        font-weight="700"
        text-anchor="middle"
        fill="#0A6B62">IQ</text>
</svg>`;
}

/**
 * Creates an SVG icon for iOS 18 Dark Mode
 * Transparent background with white foreground for glass effect
 */
function createDarkIconSVG(size) {
	const centre = size / 2;
	const aFontSize = Math.round(size * 0.82);
	const iqFontSize = Math.round(size * 0.17);
	const aY = Math.round(size * 0.79);
	const iqY = Math.round(size * 0.78);

	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <text x="${centre}" y="${aY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${aFontSize}"
        font-weight="800"
        text-anchor="middle"
        fill="#FFFFFF">A</text>
  <text x="${centre}" y="${iqY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${iqFontSize}"
        font-weight="700"
        text-anchor="middle"
        fill="#FFFFFF"
        opacity="0.9">IQ</text>
</svg>`;
}

/**
 * Creates an SVG icon for iOS 18 Tinted Mode
 * Dark teal on transparent - iOS converts to grayscale and applies user's tint
 */
function createTintedIconSVG(size) {
	const centre = size / 2;
	const aFontSize = Math.round(size * 0.82);
	const iqFontSize = Math.round(size * 0.17);
	const aY = Math.round(size * 0.79);
	const iqY = Math.round(size * 0.78);

	// Dark teal - iOS converts to grayscale for tinting
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <text x="${centre}" y="${aY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${aFontSize}"
        font-weight="800"
        text-anchor="middle"
        fill="#0A6B62">A</text>
  <text x="${centre}" y="${iqY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${iqFontSize}"
        font-weight="700"
        text-anchor="middle"
        fill="#0A6B62">IQ</text>
</svg>`;
}

/**
 * Creates an SVG icon designed for favicons, with minimal padding
 * Maximizes usage of the canvas for best visibility at small sizes (16px, 32px)
 */
function createFaviconSVG(size) {
	const centre = size / 2;
	// Maximize size - scale everything up
	const aFontSize = Math.round(size * 1.15);  // Increased to 115% size
	const iqFontSize = Math.round(size * 0.24); // Proportionally larger IQ
	const aY = Math.round(size * 0.98);         // Push down further
	const iqY = Math.round(size * 0.95);        // Push down further

	// Teal brand colour (#0A6B62)
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <text x="${centre}" y="${aY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${aFontSize}"
        font-weight="800"
        text-anchor="middle"
        fill="#0A6B62">A</text>
  <text x="${centre}" y="${iqY}"
        font-family="Inter, Arial, sans-serif"
        font-size="${iqFontSize}"
        font-weight="700"
        text-anchor="middle"
        fill="#0A6B62">IQ</text>
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

async function generateAppleIcon(filename, size) {
	const svg = createAppleIconSVG(size);
	const outputPath = join(__dirname, filename);

	// Ensure transparent background is preserved
	await sharp(Buffer.from(svg))
		.resize(size, size)
		.png({ compressionLevel: 9 })
		.toFile(outputPath);

	console.log(`✓ Generated ${filename} (${size}×${size}) [transparent]`);
}

async function generateDarkIcon(filename, size) {
	const svg = createDarkIconSVG(size);
	const outputPath = join(__dirname, filename);

	await sharp(Buffer.from(svg))
		.resize(size, size)
		.png({ compressionLevel: 9 })
		.toFile(outputPath);

	console.log(`✓ Generated ${filename} (${size}×${size}) [dark]`);
}

async function generateTintedIcon(filename, size) {
	const svg = createTintedIconSVG(size);
	const outputPath = join(__dirname, filename);

	await sharp(Buffer.from(svg))
		.resize(size, size)
		.png({ compressionLevel: 9 })
		.toFile(outputPath);

	console.log(`✓ Generated ${filename} (${size}×${size}) [tinted/grayscale]`);
}

async function generateFavicon(filename, size) {
	const svg = createFaviconSVG(size);
	const outputPath = join(__dirname, filename);

	// Favicons usually need maximum clarity
	await sharp(Buffer.from(svg))
		.resize(size, size)
		.png({ compressionLevel: 9 })
		.toFile(outputPath);

	console.log(`✓ Generated ${filename} (${size}×${size}) [maximized]`);
}

// ... (keep creating functions same, just updated colors will apply via constants)

async function writeSVG(filename, svgContent) {
	const outputPath = join(__dirname, filename);
	writeFileSync(outputPath, svgContent);
	console.log(`✓ Saved source SVG: ${filename}`);
}

async function main() {
	console.log('Generating NethAddress icons...\n');

	try {
		// 1. Standard PWA Icons (Teal Gradient Background)
		console.log('Standard icons (Teal Gradient):');
		const iconSVG = createIconSVG(512); // Master SVG
		await writeSVG('icon.svg', iconSVG); // Save source

		await generateIcon('icon-192.png', 192);
		await generateIcon('icon-512.png', 512);

		// 2. Maskable Icons (Teal Gradient)
		console.log('\nAndroid adaptive icons:');
		const maskableSVG = createIconSVG(512, true); // Master Maskable SVG
		await writeSVG('icon-maskable.svg', maskableSVG); // Save source

		await generateIcon('icon-maskable-192.png', 192, true);
		await generateIcon('icon-maskable-512.png', 512, true);

		// 3. Apple touch icons (Transparent/Teal for "Liquid" look)
		console.log('\nApple touch icons (Transparent - iOS Liquid):');
		await generateAppleIcon('apple-touch-icon-152.png', 152);
		await generateAppleIcon('apple-touch-icon-167.png', 167);
		await generateAppleIcon('apple-touch-icon-180.png', 180);
		// Note: We generally don't need to save these SVGs as they are derived, but we could.

		// 4. Favicons (Teal on Transparent)
		console.log('\nFavicons:');
		const faviconSVG = createFaviconSVG(512); // Master Favicon SVG
		await writeSVG('favicon.svg', faviconSVG); // Update the SVG file served to browsers!

		await generateFavicon('favicon-32.png', 32);
		await generateFavicon('favicon-16.png', 16);

		// 5. Variants
		await generateDarkIcon('apple-touch-icon-dark.svg', 512); // Just testing/saving if needed?
		// Actually generateDarkIcon outputs PNG. Let's just run the PNG gens.
		await generateDarkIcon('apple-touch-icon-dark-180.png', 180);
		await generateTintedIcon('apple-touch-icon-tinted-180.png', 180);

		console.log('\n✅ All icons generated successfully!');
		console.log('Icons are ready in: frontend/static/icons/');
	} catch (error) {
		console.error('Error generating icons:', error);
		process.exit(1);
	}
}

main();
