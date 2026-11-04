// The pose worker: it poses the raid test's crowd (its raiders and their effects) off the page's
// thread, into memory shared with the page (poseClient.ts).
//
// Each model it poses has two sets of bone matrices in shared memory. The page reads one set
// while we write the other, and they swap when a frame's poses are done, so the page never reads
// a pose half written.

import * as THREE from 'three';
import { animData, M2 } from './m2';
import { PoseRequest, Poser, RECORD, ToWorker } from './pose';

type Posed = { poser: Poser; fdid: number; places: Float64Array[]; skinning: Float32Array[] };

const files = new Map<number, M2>();
const anims = new Map<number, Map<number, DataView>>();
const models = new Map<number, Posed>();
// The frame's requests, in memory shared with the page (it writes them before each frame message).
let requests = new Float64Array(0);

const camera = new THREE.Quaternion();
const world = new THREE.Matrix4();
const req: PoseRequest = { globalTime: 0, base: { seq: 0, time: 0 }, fade: 0, snapshot: false, overlayWeight: 0, twist: 0 };
const overlay = { seq: 0, time: 0, source: undefined as DataView | undefined };
const hands = { seq: 0, source: undefined as DataView | undefined, right: false, left: false };

// source gives a sequence's keys: in the model's file, or in its own animation file.
function source(fdid: number, m2: M2, seq: number): DataView | undefined {
	return m2.inline(seq) ? m2.data : anims.get(fdid)?.get(seq);
}

self.onmessage = (e: MessageEvent<ToWorker>) => {
	const msg = e.data;
	switch (msg.type) {
		case 'file':
			if (!files.has(msg.fdid)) files.set(msg.fdid, new M2(msg.bytes));
			return;
		case 'anim': {
			const m2 = files.get(msg.fdid);
			if (!m2) return;
			const chunked = (m2.flags & 0x200000) !== 0 || m2.skeleton > 0;
			let list = anims.get(msg.fdid);
			if (!list) anims.set(msg.fdid, (list = new Map()));
			list.set(msg.seq, animData(msg.bytes, chunked));
			return;
		}
		case 'model': {
			const m2 = files.get(msg.fdid);
			if (!m2) return;
			const poser = new Poser(m2);
			const places = msg.places.map(b => new Float64Array(b));
			const skinning = msg.skinning.map(b => new Float32Array(b));
			models.set(msg.id, { poser, fdid: msg.fdid, places, skinning });
			return;
		}
		case 'drop':
			models.delete(msg.id);
			return;
		case 'requests':
			requests = new Float64Array(msg.buffer);
			return;
		case 'frame': {
			const d = requests;
			for (let o = 0; o + RECORD <= msg.count * RECORD; o += RECORD) {
				const m = models.get(d[o]);
				if (!m) continue;
				const target = d[o + 1];
				const m2 = m.poser.m2;
				req.globalTime = d[o + 2];
				req.base.seq = d[o + 3];
				req.base.time = d[o + 4];
				req.base.source = d[o + 5] ? source(m.fdid, m2, d[o + 3]) : undefined;
				req.fade = d[o + 6];
				req.snapshot = d[o + 7] !== 0;
				if (d[o + 8] >= 0) {
					overlay.seq = d[o + 8];
					overlay.time = d[o + 9];
					overlay.source = d[o + 10] ? source(m.fdid, m2, d[o + 8]) : undefined;
					req.overlay = overlay;
				} else req.overlay = undefined;
				req.overlayWeight = d[o + 11];
				if (d[o + 12] >= 0) {
					hands.seq = d[o + 12];
					hands.source = d[o + 13] ? source(m.fdid, m2, d[o + 12]) : undefined;
					hands.right = d[o + 14] !== 0;
					hands.left = d[o + 15] !== 0;
					req.hands = hands;
				} else req.hands = undefined;
				req.twist = d[o + 16];
				if (d[o + 17]) {
					camera.fromArray(d, o + 18);
					world.fromArray(d, o + 22);
					req.camera = camera;
					req.world = world;
				} else {
					req.camera = undefined;
					req.world = undefined;
				}
				m.poser.places = m.places[target];
				m.poser.skinning = m.skinning[target];
				m.poser.pose(req);
			}
			self.postMessage({ type: 'done', frame: msg.frame });
		}
	}
};
