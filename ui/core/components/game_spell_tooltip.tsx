// A spell's tooltip as the game draws it. The name sits at the top with the rank on its right,
// then the lines with a left and a right part (like the cast time and the cooldown), and the
// description in yellow with its numbers in white. Use it with tippy's 'game-spell' theme.
export interface GameSpellTooltipConfig {
	name: string;
	rank?: string;
	lines: Array<[string, string?]>;
	// A requirement we don't meet, like a level the character hasn't reached, in red.
	unmet?: string;
	description: string;
}

export function gameSpellTooltip(config: GameSpellTooltipConfig): HTMLElement {
	return (
		<div className="game-spell-tooltip">
			<div className="game-spell-tooltip-line">
				<span className="game-spell-tooltip-name">{config.name}</span>
				{config.rank && <span className="game-spell-tooltip-rank">{config.rank}</span>}
			</div>
			{config.lines.map(([left, right]) => (
				<div className="game-spell-tooltip-line">
					<span>{left}</span>
					{right && <span>{right}</span>}
				</div>
			))}
			{config.unmet && <div className="game-spell-tooltip-unmet">{config.unmet}</div>}
			<div className="game-spell-tooltip-description">
				{config.description.split(/(\d+(?:\.\d+)?%?)/).map((text, i) => (i % 2 ? <span className="game-spell-tooltip-value">{text}</span> : text))}
			</div>
		</div>
	) as HTMLElement;
}
