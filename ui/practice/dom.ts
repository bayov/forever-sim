// Helpers for the HUD's per-frame updates.
//
// The HUD updates every frame, at up to 240 FPS. Searching the page for an element or writing a
// text or style that didn't change still costs time, and a write makes the browser check the
// page's styles and layout again. So we remember the elements we found and write only changes.

const found = new WeakMap<ParentNode, Map<string, HTMLElement | null>>();

// find gives the first element under root that matches the selector. We remember it until it
// leaves the page (the HUD rebuilds its frames when the window's size changes).
export function find(root: ParentNode, selector: string): HTMLElement {
	let m = found.get(root);
	if (!m) found.set(root, (m = new Map()));
	let e = m.get(selector);
	if (!e || !e.isConnected) {
		e = root.querySelector(selector) as HTMLElement | null;
		m.set(selector, e);
	}
	return e as HTMLElement;
}

// setText changes an element's text when it's different.
export function setText(e: Element | null, text: string) {
	if (e && e.textContent !== text) e.textContent = text;
}

const styles = new WeakMap<HTMLElement, Map<string, string>>();

// setStyle changes one of an element's styles (like "width" or "--sweep") when it's different
// from what we set last time.
export function setStyle(e: HTMLElement | null, prop: string, value: string) {
	if (!e) return;
	let m = styles.get(e);
	if (!m) styles.set(e, (m = new Map()));
	if (m.get(prop) === value) return;
	m.set(prop, value);
	e.style.setProperty(prop, value);
}
