/**
 * Icon Generator Script for NethAddress
 *
 * Generates PNG icons from SVG sources for PWA/home screen support.
 * All icons use a consistent design: Dutch flag (red/white/blue) with
 * a transparent "A" cutout, rendered as rounded squares.
 *
 * Run with: node generate-icons.mjs
 *
 * Prerequisites: npm install sharp
 */

import sharp from 'sharp';
import { writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Dutch flag colours
const FLAG_RED = '#AE1C28';
const FLAG_WHITE = '#FFFFFF';
const FLAG_BLUE = '#21468B';

// Dark mode variants (brighter for contrast)
const FLAG_RED_DARK = '#C8303C';
const FLAG_WHITE_DARK = '#F0F0F0';
const FLAG_BLUE_DARK = '#2B5BA8';

/**
 * Creates the standard icon SVG: Dutch flag with transparent A cutout, rounded rect
 * @param {number} size - Icon dimensions
 * @param {boolean} maskable - Whether to use maskable (full-bleed, no rounded corners)
 * @returns {string} SVG content
 */
function createIconSVG(size, maskable = false) {
	const cornerRadius = maskable ? 0 : Math.round(size * 0.176);
	const centre = size / 2;

	// For maskable, A needs to fit within the safe zone (~80% centre)
	// For regular, A fills almost the entire icon
	const aFontSize = maskable ? Math.round(size * 0.82) : Math.round(size * 0.9);
	const aY = maskable ? Math.round(size * 0.76) : Math.round(size * 0.80);

	const stripeHeight = Math.round(size / 3);
	const lastStripeY = stripeHeight * 2;
	const lastStripeH = size - lastStripeY;

	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <defs>
    <mask id="cutout">
      <rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="white"/>
      <text x="${centre}" y="${aY}"
            font-family="Inter, Arial, sans-serif"
            font-size="${aFontSize}" font-weight="800"
            text-anchor="middle" fill="black">A</text>
    </mask>
  </defs>
  <g mask="url(#cutout)">
    <rect x="0" y="0" width="${size}" height="${stripeHeight}" fill="${FLAG_RED}"/>
    <rect x="0" y="${stripeHeight}" width="${size}" height="${stripeHeight}" fill="${FLAG_WHITE}"/>
    <rect x="0" y="${lastStripeY}" width="${size}" height="${lastStripeH}" fill="${FLAG_BLUE}"/>
  </g>
  ${maskable ? '' : `<rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="none" stroke="#1a1a1a" stroke-width="${Math.max(2, Math.round(size * 0.012))}" opacity="0.15"/>`}
</svg>`;
}

/**
 * Creates an Apple touch icon SVG: Dutch flag with transparent A, rounded rect
 * Transparent background for iOS Liquid Glass effects
 * @param {number} size - Icon dimensions
 * @returns {string} SVG content
 */
function createAppleIconSVG(size) {
	const cornerRadius = Math.round(size * 0.176);
	const centre = size / 2;
	const aFontSize = Math.round(size * 0.9);
	const aY = Math.round(size * 0.80);
	const stripeHeight = Math.round(size / 3);
	const lastStripeY = stripeHeight * 2;
	const lastStripeH = size - lastStripeY;

	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <defs>
    <mask id="cutout">
      <rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="white"/>
      <text x="${centre}" y="${aY}"
            font-family="Inter, Arial, sans-serif"
            font-size="${aFontSize}" font-weight="800"
            text-anchor="middle" fill="black">A</text>
    </mask>
  </defs>
  <g mask="url(#cutout)">
    <rect x="0" y="0" width="${size}" height="${stripeHeight}" fill="${FLAG_RED}"/>
    <rect x="0" y="${stripeHeight}" width="${size}" height="${stripeHeight}" fill="${FLAG_WHITE}"/>
    <rect x="0" y="${lastStripeY}" width="${size}" height="${lastStripeH}" fill="${FLAG_BLUE}"/>
  </g>
  <rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="none" stroke="#1a1a1a" stroke-width="2" opacity="0.12"/>
</svg>`;
}

/**
 * Creates an Apple Dark mode icon SVG
 */
function createDarkIconSVG(size) {
	const cornerRadius = Math.round(size * 0.176);
	const centre = size / 2;
	const aFontSize = Math.round(size * 0.9);
	const aY = Math.round(size * 0.80);
	const stripeHeight = Math.round(size / 3);
	const lastStripeY = stripeHeight * 2;
	const lastStripeH = size - lastStripeY;

	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <defs>
    <mask id="cutout">
      <rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="white"/>
      <text x="${centre}" y="${aY}"
            font-family="Inter, Arial, sans-serif"
            font-size="${aFontSize}" font-weight="800"
            text-anchor="middle" fill="black">A</text>
    </mask>
  </defs>
  <g mask="url(#cutout)">
    <rect x="0" y="0" width="${size}" height="${stripeHeight}" fill="${FLAG_RED_DARK}"/>
    <rect x="0" y="${stripeHeight}" width="${size}" height="${stripeHeight}" fill="${FLAG_WHITE_DARK}"/>
    <rect x="0" y="${lastStripeY}" width="${size}" height="${lastStripeH}" fill="${FLAG_BLUE_DARK}"/>
  </g>
  <rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="none" stroke="#ffffff" stroke-width="2" opacity="0.15"/>
</svg>`;
}

/**
 * Creates an Apple Tinted mode icon SVG
 */
function createTintedIconSVG(size) {
	const cornerRadius = Math.round(size * 0.176);
	const centre = size / 2;
	const aFontSize = Math.round(size * 0.9);
	const aY = Math.round(size * 0.80);
	const stripeHeight = Math.round(size / 3);
	const lastStripeY = stripeHeight * 2;
	const lastStripeH = size - lastStripeY;

	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}">
  <defs>
    <mask id="cutout">
      <rect width="${size}" height="${size}" rx="${cornerRadius}" ry="${cornerRadius}" fill="white"/>
      <text x="${centre}" y="${aY}"
            font-family="Inter, Arial, sans-serif"
            font-size="${aFontSize}" font-weight="800"
            text-anchor="middle" fill="black">A</text>
    </mask>
  </defs>
  <g mask="url(#cutout)">
    <rect x="0" y="0" width="${size}" height="${stripeHeight}" fill="#333333" opacity="0.9"/>
    <rect x="0" y="${stripeHeight}" width="${size}" height="${stripeHeight}" fill="#333333" opacity="0.3"/>
    <rect x="0" y="${lastStripeY}" width="${size}" height="${lastStripeH}" fill="#333333" opacity="0.7"/>
  </g>
</svg>`;
}

/**
 * Creates a favicon SVG: same design as main icon, rounded rect
 */
function createFaviconSVG(size) {
	// Favicons use the same design as the main icon
	return createIconSVG(size, false);
}

async function generateIcon(filename, size, maskable = false) {
	const svg = createIconSVG(size, maskable);
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

	await sharp(Buffer.from(svg))
		.resize(size, size)
		.png({ compressionLevel: 9 })
		.toFile(outputPath);

	console.log(`✓ Generated ${filename} (${size}×${size}) [apple]`);
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

	console.log(`✓ Generated ${filename} (${size}×${size}) [tinted]`);
}

async function generateFavicon(filename, size) {
	const svg = createFaviconSVG(size);
	const outputPath = join(__dirname, filename);

	await sharp(Buffer.from(svg))
		.resize(size, size)
		.png({ compressionLevel: 9 })
		.toFile(outputPath);

	console.log(`✓ Generated ${filename} (${size}×${size}) [favicon]`);
}

async function writeSVG(filename, svgContent) {
	const outputPath = join(__dirname, filename);
	writeFileSync(outputPath, svgContent);
	console.log(`✓ Saved source SVG: ${filename}`);
}

async function main() {
	console.log('Generating NethAddress icons...\n');

	try {
		// 1. Standard PWA Icons (Dutch flag, rounded rect)
		console.log('Standard icons:');
		const iconSVG = createIconSVG(512);
		await writeSVG('icon.svg', iconSVG);

		await generateIcon('icon-192.png', 192);
		await generateIcon('icon-512.png', 512);

		// 2. Maskable Icons (full-bleed, no rounded corners)
		console.log('\nMaskable icons (Android adaptive):');
		const maskableSVG = createIconSVG(512, true);
		await writeSVG('icon-maskable.svg', maskableSVG);

		await generateIcon('icon-maskable-192.png', 192, true);
		await generateIcon('icon-maskable-512.png', 512, true);

		// 3. Apple touch icons (rounded rect, Dutch flag)
		console.log('\nApple touch icons:');
		const appleSVG = createAppleIconSVG(180);
		await writeSVG('apple-touch-icon.svg', appleSVG);

		await generateAppleIcon('apple-touch-icon-152.png', 152);
		await generateAppleIcon('apple-touch-icon-167.png', 167);
		await generateAppleIcon('apple-touch-icon-180.png', 180);

		// 4. Apple Dark mode icons
		console.log('\nApple dark mode icons:');
		const darkSVG = createDarkIconSVG(180);
		await writeSVG('apple-touch-icon-dark.svg', darkSVG);

		await generateDarkIcon('apple-touch-icon-dark-152.png', 152);
		await generateDarkIcon('apple-touch-icon-dark-167.png', 167);
		await generateDarkIcon('apple-touch-icon-dark-180.png', 180);

		// 5. Apple Tinted mode icons
		console.log('\nApple tinted mode icons:');
		const tintedSVG = createTintedIconSVG(180);
		await writeSVG('apple-touch-icon-tinted.svg', tintedSVG);

		await generateTintedIcon('apple-touch-icon-tinted-152.png', 152);
		await generateTintedIcon('apple-touch-icon-tinted-167.png', 167);
		await generateTintedIcon('apple-touch-icon-tinted-180.png', 180);

		// 6. Favicons (same design, rounded rect)
		console.log('\nFavicons:');
		const faviconSVG = createFaviconSVG(512);
		await writeSVG('favicon.svg', faviconSVG);

		await generateFavicon('favicon-32.png', 32);
		await generateFavicon('favicon-16.png', 16);

		console.log('\n✅ All icons generated successfully!');
		console.log('Icons are ready in: frontend/static/icons/');
	} catch (error) {
		console.error('Error generating icons:', error);
		process.exit(1);
	}
}

main();
