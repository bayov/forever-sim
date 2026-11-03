// Movement and the camera, as WoW does them (docs/practice_room.md, "Movement" and "Camera").
//
// Your own movement is the page's: it moves you at once, and tells the session where you are and
// how you move each time that changes, as WoW trusts the client's movement.

import { CameraState } from './scene';

export const speeds = { run: 7, walk: 2.5, backpedal: 4.5, turn: Math.PI };
const jumpTime = 0.83;
const jumpHeight = 1.64;

export type MoveState = {
	pos: [number, number];
	facing: number;
	// velocity in yards per second, turnRate in radians per second
	velocity: [number, number];
	turnRate: number;
	jumpStart?: number; // wall ms
};

export type Held = {
	forward: boolean;
	back: boolean;
	strafeLeft: boolean;
	strafeRight: boolean;
	turnLeft: boolean;
	turnRight: boolean;
};

export type CameraOptions = {
	yawSpeed: number; // cameraYawMoveSpeed, 180 by default
	pitchSpeed: number; // cameraPitchMoveSpeed, 90
	smoothStyle: number; // cameraSmoothStyle, 4 by default
	yawSmoothSpeed: number; // cameraYawSmoothSpeed, 180 degrees a second
	maxDistance: number; // 15 yards times cameraDistanceMaxZoomFactor
	zoomStep: number;
};

export class Mover {
	held: Held = { forward: false, back: false, strafeLeft: false, strafeRight: false, turnLeft: false, turnRight: false };
	autorun = false;
	// speedFactor is 1.4 in Ghost Wolf form.
	speedFactor = 1;
	leftDown = false;
	rightDown = false;
	readonly move: MoveState;
	readonly cam: CameraState;
	// Where the wheel is taking the camera, while it glides there.
	private zoomTarget?: number;
	// The last movement the session heard about, to send only changes.
	private sent = { vx: 0, vy: 0, turn: 0, facing: 0 };
	private lastFacingSent = 0;

	constructor(
		start: [number, number],
		facing: number,
		readonly options: CameraOptions,
		readonly send: (m: { pos: [number, number]; velocity: [number, number]; facing: number; turnRate: number }) => void,
		readonly onJump: () => void,
	) {
		this.move = { pos: [...start] as [number, number], facing, velocity: [0, 0], turnRate: 0 };
		this.cam = { yaw: facing, pitch: (10 * Math.PI) / 180, distance: 5.55 };
		this.sent.facing = facing;
	}

	// The turn keys strafe while the right button is held, and both buttons run forward.
	private intent() {
		const h = this.held;
		const rmb = this.rightDown;
		let forward = h.forward || this.autorun || (this.leftDown && this.rightDown);
		let back = h.back;
		if (forward && back) forward = back = false;
		let strafe = 0;
		if (h.strafeLeft || (rmb && h.turnLeft)) strafe += 1;
		if (h.strafeRight || (rmb && h.turnRight)) strafe -= 1;
		let turn = 0;
		if (!rmb) {
			if (h.turnLeft) turn += 1;
			if (h.turnRight) turn -= 1;
		}
		return { forward, back, strafe, turn };
	}

	velocityFor(facing: number): [number, number] {
		const { forward, back, strafe } = this.intent();
		let dx = 0;
		let dy = 0;
		if (forward) dx += 1;
		if (back) dx -= 1;
		dy += strafe;
		if (dx === 0 && dy === 0) return [0, 0];
		const len = Math.hypot(dx, dy);
		const speed = (back ? speeds.backpedal : speeds.run) * this.speedFactor;
		const lx = (dx / len) * speed;
		const ly = (dy / len) * speed;
		// Local (forward, left) to world.
		const c = Math.cos(facing);
		const s = Math.sin(facing);
		return [lx * c - ly * s, lx * s + ly * c];
	}

	airborne(now: number): boolean {
		return this.move.jumpStart !== undefined && now - this.move.jumpStart < jumpTime * 1000;
	}

	height(now: number): number {
		if (!this.airborne(now) || this.move.jumpStart === undefined) return 0;
		const t = (now - this.move.jumpStart) / 1000;
		const k = t / jumpTime;
		return 4 * jumpHeight * k * (1 - k);
	}

	jump(now: number) {
		if (this.airborne(now)) return;
		this.move.jumpStart = now;
		this.onJump();
	}

	// step moves you by dt seconds and sends the session any change in how you move.
	step(dt: number, now: number) {
		const m = this.move;
		const { turn } = this.intent();
		m.turnRate = turn * speeds.turn;
		m.facing = normalize(m.facing + m.turnRate * dt);
		// A jump keeps its direction, but you can still turn.
		if (!this.airborne(now)) {
			m.velocity = this.velocityFor(m.facing);
		}
		m.pos = [m.pos[0] + m.velocity[0] * dt, m.pos[1] + m.velocity[1] * dt];

		// The camera swings back behind you while you move (cameraSmoothStyle 1 and 4), unless a
		// mouse button holds it.
		const moving = m.velocity[0] !== 0 || m.velocity[1] !== 0 || m.turnRate !== 0;
		const style = this.options.smoothStyle;
		if (!this.leftDown && !this.rightDown && (style === 2 || ((style === 1 || style === 4) && moving))) {
			const diff = angleDiff(m.facing, this.cam.yaw);
			const maxStep = ((this.options.yawSmoothSpeed * Math.PI) / 180) * dt;
			this.cam.yaw = normalize(this.cam.yaw + Math.max(-maxStep, Math.min(maxStep, diff)));
		}
		if (this.rightDown) this.cam.yaw = m.facing;

		this.maybeSend(now);
	}

	private maybeSend(now: number) {
		const m = this.move;
		const s = this.sent;
		const changed = m.velocity[0] !== s.vx || m.velocity[1] !== s.vy || m.turnRate !== s.turn;
		// Mouse turning changes facing with no turn rate: we send it every 50ms while it changes.
		const facingDrift = Math.abs(angleDiff(m.facing, s.facing + s.turn * ((now - this.lastFacingSent) / 1000)));
		if (changed || (facingDrift > 0.01 && now - this.lastFacingSent >= 50)) {
			s.vx = m.velocity[0];
			s.vy = m.velocity[1];
			s.turn = m.turnRate;
			s.facing = m.facing;
			this.lastFacingSent = now;
			this.send({ pos: [m.pos[0], m.pos[1]], velocity: [m.velocity[0], m.velocity[1]], facing: m.facing, turnRate: m.turnRate });
		}
	}

	// Mouse look: left drag orbits the camera, right drag turns you to where the camera looks.
	look(dx: number, dy: number) {
		const yawPerPixel = ((this.options.yawSpeed / 180) * 0.3 * Math.PI) / 180;
		const pitchPerPixel = ((this.options.pitchSpeed / 90) * 0.3 * Math.PI) / 180;
		this.cam.yaw = normalize(this.cam.yaw - dx * yawPerPixel);
		this.cam.pitch = Math.max(-1.4, Math.min(1.5, this.cam.pitch + dy * pitchPerPixel));
		if (this.rightDown) this.move.facing = this.cam.yaw;
	}

	// zoom moves the camera in or out by wheel steps. The camera glides there (updateZoom), like
	// the game's.
	zoom(steps: number) {
		const from = this.zoomTarget ?? this.cam.distance;
		this.zoomTarget = Math.max(0.5, Math.min(this.options.maxDistance, from + steps * this.options.zoomStep));
	}

	// updateZoom moves the camera a frame's share of the way to where the wheel sent it. It covers
	// most of the way in about a tenth of a second.
	updateZoom(dt: number) {
		const target = this.zoomTarget;
		if (target === undefined) return;
		const d = this.cam.distance;
		const next = d + (target - d) * Math.min(1, dt * 15);
		if (Math.abs(target - next) < 0.005) {
			this.cam.distance = target;
			this.zoomTarget = undefined;
		} else this.cam.distance = next;
	}
}

export function normalize(a: number): number {
	a %= 2 * Math.PI;
	return a < 0 ? a + 2 * Math.PI : a;
}

export function angleDiff(to: number, from: number): number {
	let d = (to - from) % (2 * Math.PI);
	if (d > Math.PI) d -= 2 * Math.PI;
	if (d < -Math.PI) d += 2 * Math.PI;
	return d;
}
