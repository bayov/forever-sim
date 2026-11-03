// The room's own mouse cursor, so Chrome's "to show your cursor, press Esc" notice shows once.
//
// Chrome shows that notice every time a page hides the cursor (pointer lock). The game hides the
// cursor each time you hold a mouse button to turn the camera, so the notice kept popping up. So
// on your first click in the room we lock the mouse for good and draw the cursor ourselves: we
// move it by the mouse's movement and send the page the mouse events a real cursor would make
// (moves, hovers, presses, clicks and the wheel) at its spot. Turning the camera then only hides
// our cursor.
//
// The browser lets go of the lock when you press Esc (outside full screen, where we lock Esc) or
// Alt+Tab away. Then the real cursor is back until your next click in the room.

const hoverClass = 'v-hover';

export class Pointer {
	x = 0;
	y = 0;
	// While you turn the camera the cursor hides and stays put.
	private frozen = false;
	private readonly cursor = document.createElement('div');
	private cursorUrl = '';
	private over: Element | null = null;
	private pressed = new Map<number, Element>();
	private select?: HTMLElement;
	// The page's pixels per unit of the mouse's movement, learned from the real cursor, so ours
	// moves at the same speed.
	private scale = 1;
	private learned = { client: 0, movement: 0 };
	private last?: { x: number; y: number };
	private readonly listening = new AbortController();

	constructor(readonly root: HTMLElement) {
		this.cursor.className = 'v-cursor hidden';
		document.body.append(this.cursor);
		// Our listeners on the window's capture phase see each real event first. While the mouse is
		// locked we stop the real one (its spot is meaningless) and send our own instead.
		const capture = { capture: true, passive: false, signal: this.listening.signal };
		window.addEventListener('mousedown', e => this.onDown(e), capture);
		window.addEventListener('mouseup', e => this.onUp(e), capture);
		window.addEventListener('mousemove', e => this.onMove(e), capture);
		window.addEventListener('wheel', e => this.onWheel(e), capture);
		window.addEventListener('contextmenu', e => this.active && e.isTrusted && (e.preventDefault(), e.stopImmediatePropagation()), capture);
		document.addEventListener('pointerlockchange', () => this.onLockChange(), { signal: this.listening.signal });
	}

	// dispose stops the cursor, when you leave the room.
	dispose() {
		this.listening.abort();
		this.cursor.remove();
		this.closeSelect();
	}

	// active is whether the mouse is locked to the room and our cursor stands in for it.
	get active(): boolean {
		return document.pointerLockElement === this.root;
	}

	// freeze hides the cursor while you turn the camera.
	freeze() {
		this.frozen = true;
		this.cursor.classList.add('hidden');
	}

	// thaw shows the cursor again where the turn started, like the game does.
	thaw(x: number, y: number) {
		this.frozen = false;
		this.x = x;
		this.y = y;
		if (this.active) {
			this.cursor.classList.remove('hidden');
			this.hover();
		}
	}

	private lock(e: MouseEvent) {
		if (!this.root.isConnected || !this.root.contains(e.target as Node)) return;
		this.x = e.clientX;
		this.y = e.clientY;
		// The lock needs the click in progress, and Chrome refuses it for a moment after the last one
		// ended. Then the next click tries again.
		const p = this.root.requestPointerLock?.() as Promise<void> | undefined;
		p?.catch?.(() => {});
	}

	private onLockChange() {
		if (this.active) {
			this.cursor.classList.toggle('hidden', this.frozen);
			this.hover();
			return;
		}
		this.cursor.classList.add('hidden');
		this.setOver(null);
		this.closeSelect();
		this.frozen = false;
	}

	private onDown(e: MouseEvent) {
		if (!e.isTrusted) return;
		if (!this.active) {
			this.lock(e);
			return;
		}
		e.preventDefault();
		e.stopImmediatePropagation();
		const target = this.hover();
		if (!target) return;
		this.pressed.set(e.button, target);
		if (this.select && !this.select.contains(target)) this.closeSelect();
		// Focus follows the press, like it would for a real click on a field.
		const focusable = (target as HTMLElement).closest?.('input, select, textarea, button') as HTMLElement | null;
		if (focusable) focusable.focus();
		else (document.activeElement as HTMLElement | null)?.blur?.();
		this.send('mousedown', target, e);
		if (e.button === 0 && target instanceof HTMLSelectElement) this.openSelect(target);
	}

	private onUp(e: MouseEvent) {
		if (!e.isTrusted || !this.active) return;
		e.preventDefault();
		e.stopImmediatePropagation();
		const target = this.frozen ? (this.pressed.get(e.button) ?? this.over) : this.hover();
		if (!target) return;
		this.send('mouseup', target, e);
		const down = this.pressed.get(e.button);
		this.pressed.delete(e.button);
		if (e.button === 0 && down === target) this.send('click', target, e);
	}

	private onMove(e: MouseEvent) {
		if (!e.isTrusted) return;
		if (!this.active) {
			this.learn(e);
			return;
		}
		e.stopImmediatePropagation();
		const dx = e.movementX * this.scale;
		const dy = e.movementY * this.scale;
		if (!this.frozen) {
			this.x = Math.max(0, Math.min(innerWidth - 1, this.x + dx));
			this.y = Math.max(0, Math.min(innerHeight - 1, this.y + dy));
		}
		const target = this.frozen ? (this.pressed.values().next().value ?? this.over) : this.hover();
		if (!target) return;
		this.send('mousemove', target, e, dx, dy);
		// The move may have changed the cursor (the sword over the dummy).
		if (!this.frozen) this.drawCursor(target);
	}

	private onWheel(e: WheelEvent) {
		if (!e.isTrusted || !this.active) return;
		e.preventDefault();
		e.stopImmediatePropagation();
		const target = this.over;
		if (!target) return;
		target.dispatchEvent(new WheelEvent('wheel', { ...this.init(e), deltaX: e.deltaX, deltaY: e.deltaY, deltaZ: e.deltaZ, deltaMode: e.deltaMode }));
	}

	// learn compares the real cursor's moves with the mouse's movement.
	private learn(e: MouseEvent) {
		const last = this.last;
		this.last = { x: e.clientX, y: e.clientY };
		if (!last || (e.movementX === 0 && e.movementY === 0)) return;
		this.learned.client += Math.abs(e.clientX - last.x) + Math.abs(e.clientY - last.y);
		this.learned.movement += Math.abs(e.movementX) + Math.abs(e.movementY);
		if (this.learned.movement > 300) this.scale = Math.max(0.5, Math.min(3, this.learned.client / this.learned.movement));
	}

	// hover finds what's under the cursor, sends the enter and leave events when that changes, and
	// draws the cursor that element asks for (its CSS cursor).
	private hover(): Element | null {
		const target = document.elementFromPoint(this.x, this.y);
		this.setOver(target);
		this.cursor.style.transform = `translate(${Math.round(this.x)}px, ${Math.round(this.y)}px)`;
		this.drawCursor(target);
		return target;
	}

	// drawCursor shows the cursor an element asks for. Outside the room (our dropdown lists) the
	// cursor stays as it was.
	private drawCursor(target: Element | null) {
		if (!target) return;
		const url = /url\("?([^")]+)"?\)/.exec(getComputedStyle(target).cursor)?.[1];
		if (url && url !== this.cursorUrl) {
			this.cursorUrl = url;
			this.cursor.style.backgroundImage = `url("${url}")`;
		}
	}

	private setOver(target: Element | null) {
		const old = this.over;
		if (old === target) return;
		this.over = target;
		const chain = (e: Element | null) => {
			const out: Element[] = [];
			for (let x = e; x; x = x.parentElement) out.push(x);
			return out;
		};
		const before = chain(old);
		const after = chain(target);
		const fake = { clientX: this.x, clientY: this.y, bubbles: true, cancelable: true, view: window };
		if (old) old.dispatchEvent(new MouseEvent('mouseout', { ...fake, relatedTarget: target }));
		for (const e of before) {
			if (after.includes(e)) continue;
			e.classList.remove(hoverClass);
			e.dispatchEvent(new MouseEvent('mouseleave', { ...fake, bubbles: false, relatedTarget: target }));
		}
		if (target) target.dispatchEvent(new MouseEvent('mouseover', { ...fake, relatedTarget: old }));
		for (const e of [...after].reverse()) {
			if (before.includes(e)) continue;
			e.classList.add(hoverClass);
			e.dispatchEvent(new MouseEvent('mouseenter', { ...fake, bubbles: false, relatedTarget: old }));
		}
	}

	private init(e: MouseEvent, movementX = 0, movementY = 0): MouseEventInit {
		return {
			bubbles: true,
			cancelable: true,
			view: window,
			clientX: this.x,
			clientY: this.y,
			screenX: this.x,
			screenY: this.y,
			button: e.button,
			buttons: e.buttons,
			altKey: e.altKey,
			ctrlKey: e.ctrlKey,
			shiftKey: e.shiftKey,
			metaKey: e.metaKey,
			movementX,
			movementY,
		};
	}

	private send(type: string, target: Element, e: MouseEvent, movementX = 0, movementY = 0) {
		target.dispatchEvent(new MouseEvent(type, this.init(e, movementX, movementY)));
	}

	// openSelect shows a dropdown's choices as a list we draw, since Chrome's own list needs the
	// real cursor.
	private openSelect(select: HTMLSelectElement) {
		this.closeSelect();
		const r = select.getBoundingClientRect();
		const list = document.createElement('div');
		list.className = 'v-select';
		list.style.left = `${r.left}px`;
		list.style.top = `${r.bottom}px`;
		list.style.minWidth = `${r.width}px`;
		list.style.font = getComputedStyle(select).font;
		for (const o of select.options) {
			const item = document.createElement('div');
			item.className = `v-option${o.selected ? ' selected' : ''}`;
			item.textContent = o.textContent;
			item.addEventListener('mousedown', ev => {
				ev.stopPropagation();
				this.closeSelect();
				if (select.value !== o.value) {
					select.value = o.value;
					select.dispatchEvent(new Event('change', { bubbles: true }));
				}
			});
			list.append(item);
		}
		document.body.append(list);
		this.select = list;
	}

	private closeSelect() {
		this.select?.remove();
		this.select = undefined;
	}
}
