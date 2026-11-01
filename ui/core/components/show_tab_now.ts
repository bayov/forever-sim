import { Tab } from 'bootstrap';

// Opens the tab of a tab link at once, without its fade in.
//
// Bootstrap fades a tab's pane in. It puts the pane in place, then shows it (adds "show") from a
// timer when the fade ends. On page load, when we open the tab the URL names, the page is still
// busy building itself and loading its data, so that timer fires late. Then the tab stays empty
// for a moment (we measured 250 to 400ms, and seconds in the raid sim).
export const showTabNow = (link: HTMLElement) => {
	Tab.getOrCreateInstance(link).show();
	const target = link.dataset.bsTarget;
	if (target) document.querySelector(target)?.classList.add('show');
};
