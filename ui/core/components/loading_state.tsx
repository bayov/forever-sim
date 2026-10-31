// What a page or tab shows while it reads data it already has, like the last run's results
// after a page refresh. Without it the page shows "No results yet" for a second or two and
// then the results pop in.
//
// It's a shimmering outline of a results page: a row of number tiles over a table. So the
// page keeps its shape while it loads, and the results fill in where the outline was.
const ROW_WIDTHS = [72, 58, 64, 46, 52, 38];

const bar = (className: string, width?: number) => <span className={`ls-bar ${className}`} style={width ? { width: `${width}%` } : {}} />;

export const loadingState = (text: string) => (
	<div className="loading-state" attributes={{ role: 'status' }}>
		<div className="ls-tiles">
			{[0, 1, 2, 3].map(() => (
				<div className="ls-tile">
					{bar('ls-bar-label', 45)}
					{bar('ls-bar-value', 70)}
				</div>
			))}
		</div>
		<div className="ls-table">
			{bar('ls-bar-heading', 22)}
			{ROW_WIDTHS.map(width => (
				<div className="ls-row">
					<span className="ls-bar ls-icon" />
					{bar('ls-bar-name', width * 0.6)}
					<span className="ls-row-fill" />
					{bar('ls-bar-meter', width * 0.4)}
					{bar('ls-bar-number', 10)}
				</div>
			))}
		</div>
		<div className="ls-label">
			<span className="ls-pulse" />
			{text}
		</div>
	</div>
);
