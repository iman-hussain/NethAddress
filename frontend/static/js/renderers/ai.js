/**
 * AI Summary Renderer
 */

export function renderGeminiAI(data) {
	if (!data) return '<div class="notification is-warning">No AI data available</div>';

	// If still generating (should happen via skeleton, but just in case)
	if (!data.Generated && !data.Summary) {
		return `
            <div class="content has-text-centered">
                <p><i class="fas fa-spinner fa-spin"></i> Generating insights...</p>
            </div>
        `;
	}

	// Convert newlines to paragraphs/breaks
	let formattedSummary = data.Summary
		.split('\n\n').map(para => `<p class="mb-3">${para}</p>`).join('')
		.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>'); // Basic bold support

	return `
        <div class="content">
            <div class="ai-summary-text">
                ${formattedSummary}
            </div>

            <div class="field is-grouped is-grouped-multiline mt-4">
                <div class="control">
                    <span class="tag is-info is-light">
                        <span class="icon is-small mr-1"><i class="fas fa-robot"></i></span>
                        ${data.Model || 'Gemini 2.0 Flash'}
                    </span>
                </div>
                <div class="control">
                    <span class="tag is-light">
                        <span class="icon is-small mr-1"><i class="far fa-clock"></i></span>
                        ${new Date(data.GeneratedAt).toLocaleTimeString()}
                    </span>
                </div>
            </div>
        </div>
    `;
}
