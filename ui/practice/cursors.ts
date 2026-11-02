// The game's mouse cursors, from the client's Interface\Cursor files: the gauntlet that points,
// and the sword over an enemy you can attack. They become CSS variables that styles.css uses.

import { decodeBLP } from './wow/blp';
import { file } from './wow/files';

// The cursors by their file ids (Interface\Cursor\Point.blp and Attack.blp). Their tip is the
// top left pixel, like the game's.
const cursors: Record<string, number> = { point: 131028, attack: 131013 };

export async function loadCursors(root: HTMLElement) {
	await Promise.all(
		Object.entries(cursors).map(async ([name, fdid]) => {
			try {
				const img = decodeBLP(await file(fdid));
				const canvas = document.createElement('canvas');
				canvas.width = img.width;
				canvas.height = img.height;
				canvas.getContext('2d')!.putImageData(new ImageData(img.rgba, img.width, img.height), 0, 0);
				root.style.setProperty(`--wow-${name}`, `url(${canvas.toDataURL()}) 0 0, auto`);
			} catch (e) {
				console.warn('cursor', name, e);
			}
		}),
	);
}
