import tippy, { Instance, Plugin } from 'tippy.js';

// Anything in a tooltip we can click or type in.
const CLICKABLE = 'a[href], button, input, select, textarea, summary, [role="button"]';

// A tooltip with a link or a button in it stays up while the mouse is on it, so we can click them.
//
// tippy hides a tooltip as soon as the mouse leaves what it belongs to, unless it's interactive.
// Most of our tooltips are plain text, and those should go away right away. So we look at what
// the tooltip holds each time it shows or changes, and make it interactive when it has something
// to click, like the wowhead link of a spell or the "Open in Specs" button in the report.
//
// tippy puts an interactive tooltip next to what it belongs to, instead of in body. A box that
// scrolls or clips (like the sidebar or the timeline) would cut it off there, so it stays in
// body unless it was told where to go.
export const interactiveWhenClickable: Plugin = {
	name: 'interactiveWhenClickable',
	fn(instance: Instance) {
		const check = () => {
			if (instance.props.interactive || !instance.popper.querySelector(CLICKABLE)) return;
			instance.setProps({
				interactive: true,
				...(instance.props.appendTo === tippy.defaultProps.appendTo ? { appendTo: () => document.body } : {}),
			});
		};
		return { onMount: check, onAfterUpdate: check };
	},
};
