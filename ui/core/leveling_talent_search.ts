import { countsOf, countsToTalentsString, pickProblems, TalentShape } from './leveling_plan';
import { LevelingTalentPick } from './proto/ui';

// The talent route search of the leveling plan (docs/leveling_plan.md, Talent search).
//
// A route is the order we put our talent points in from the plan's first level, one pick a
// point. We score a route by the hours of all its levels together, never one level at a
// time. Taking the talent that's best at level 10 can leave us slower over the whole way, like
// a point that does the most at 10 but keeps us out of a row we want at 20. So a change to the
// route is kept only when the plan's total hours go down.
//
// The search starts from many routes: the plan's own build, one that fills each tree from its
// top row down, and one for each deeper talent that goes straight for it (like Stormstrike at
// 25), with the fewest points before it that its row needs. A change at a time can't get from
// one of those to another, because the points in between are worse, so we score them all and
// search from the best few. From each start we keep making the change to the route that saves
// the most hours over all the levels, until no change saves any. The changes are:
// - Swapping two picks up to SWAP_WINDOW points apart, so a talent comes earlier or later.
// - Taking out the last point of a talent and putting a point in another talent instead, in its
//   place or at the end of the route.
//
// Each level is simmed on its own fixed sessions, so two routes with the same talents at a
// level get the same hours there, and we sim each level and talents only once. That keeps the
// search to a few hundred sims of a level for each start.
//
// A respec at a milestone level (like 30) is the same search over the levels from it on, with
// all our points free to go anywhere at that level. We compare it with the best route without
// a respec over the same levels, and the player decides whether the hours are worth the gold.

export interface TalentRouteSpace {
	shape: TalentShape;
	// The levels we add the hours up over, lowest first.
	levels: number[];
	// The talent points we have at a level.
	pointsAt: (level: number) => number;
	// Whether we may put points in a talent. The sim gives nothing for some talents, like
	// Improved Ghost Wolf, so a point there would be wasted.
	allowed: (tree: number, talent: number) => boolean;
}

export interface LevelHours {
	hours: number;
	// The share of the hours that is noise.
	relativeError: number;
}

// Sims each level with its talents string, and gives back the hours of each.
export type SimLevels = (jobs: Array<{ level: number; talents: string }>) => Promise<LevelHours[]>;

export interface RouteSearchProgress {
	stage: string;
	sims: number;
	bestHours: number;
}

export interface RouteResult {
	picks: LevelingTalentPick[];
	hours: number;
	// The hours of each level of the levels searched over.
	levels: Array<{ level: number } & LevelHours>;
}

// How far apart two picks can be for a swap.
const SWAP_WINDOW = 3;
// How many of the point moves that look best at the screening levels we sim over all levels.
const SET_MOVES_KEPT = 6;
// The least share of the hours a change must save to be kept. Below it we'd follow the noise of
// the sessions.
const MIN_GAIN = 0.0005;
// How many starts we search from, the best ones by their own hours.
const STARTS_SEARCHED = 3;

export class TalentRouteSearch {
	private readonly cache = new Map<string, LevelHours>();
	sims = 0;
	private stopped = false;

	constructor(
		private readonly space: TalentRouteSpace,
		private readonly simLevels: SimLevels,
		private readonly onProgress: (progress: RouteSearchProgress) => void = () => {},
	) {}

	stop() {
		this.stopped = true;
	}

	get isStopped() {
		return this.stopped;
	}

	// The number of points a route has, the points of the last level.
	private get routeLength(): number {
		return this.space.pointsAt(this.space.levels[this.space.levels.length - 1]);
	}

	talentsAt(route: LevelingTalentPick[], level: number): string {
		return countsToTalentsString(countsOf(route.slice(0, this.space.pointsAt(level)), this.space.shape));
	}

	private key(level: number, talents: string) {
		return `${level}|${talents}`;
	}

	// The hours of each route over the levels. We sim the levels and talents we haven't simmed
	// yet, all in one batch.
	private async score(routes: LevelingTalentPick[][], levels: number[]): Promise<number[]> {
		await this.simMissing(routes.flatMap(route => levels.map(level => ({ level, talents: this.talentsAt(route, level) }))));
		return routes.map(route => levels.reduce((sum, level) => sum + this.hoursAt(level, this.talentsAt(route, level)), 0));
	}

	// The hours of a level we simmed, or Infinity when we stopped before it was.
	private hoursAt(level: number, talents: string): number {
		return this.cache.get(this.key(level, talents))?.hours ?? Infinity;
	}

	private async simMissing(jobs: Array<{ level: number; talents: string }>) {
		const missing = new Map<string, { level: number; talents: string }>();
		for (const job of jobs) {
			const key = this.key(job.level, job.talents);
			if (!this.cache.has(key)) missing.set(key, job);
		}
		if (!missing.size || this.stopped) return;
		const list = [...missing.values()];
		const results = await this.simLevels(list);
		list.forEach((job, i) => {
			if (results[i]) this.cache.set(this.key(job.level, job.talents), results[i]);
		});
		this.sims += list.length;
	}

	async result(picks: LevelingTalentPick[], levels = this.space.levels): Promise<RouteResult> {
		const [hours] = await this.score([picks], levels);
		return {
			picks,
			hours,
			levels: levels.map(level => ({
				level,
				...(this.cache.get(this.key(level, this.talentsAt(picks, level))) ?? { hours: Infinity, relativeError: 0 }),
			})),
		};
	}

	private valid(route: LevelingTalentPick[]): boolean {
		return pickProblems(route, this.space.shape).every(problem => problem == null);
	}

	// A route of routeLength points that starts with the picks. We fill what's missing from
	// the top row down, in the tree we name or the one with the most points, and then the others.
	extend(picks: LevelingTalentPick[], tree?: number): LevelingTalentPick[] {
		const shape = this.space.shape;
		const route = picks.slice(0, this.routeLength);
		const counts = countsOf(route, shape);
		const first = tree ?? counts.map((c, t) => ({ t, n: c.reduce((a, b) => a + b, 0) })).sort((a, b) => b.n - a.n)[0].t;
		const trees = [first, ...shape.trees.map((_, t) => t).filter(t => t != first)];
		while (route.length < this.routeLength) {
			let next: LevelingTalentPick | null = null;
			for (const t of trees) {
				const talents = shape.trees[t].talents
					.map((config, i) => ({ config, i }))
					.filter(({ i }) => this.space.allowed(t, i))
					.sort((a, b) => a.config.row - b.config.row || a.i - b.i);
				for (const { i } of talents) {
					const pick = LevelingTalentPick.create({ tree: t, talent: i });
					if (this.valid([...route, pick])) {
						next = pick;
						break;
					}
				}
				if (next) break;
			}
			if (!next) break;
			route.push(next);
		}
		return route;
	}

	// A route that goes straight for a deeper talent: the points its row needs in its tree (the
	// talents the start has there first), then the talent it needs, then all its ranks, and the
	// rest like extend. Null when the route has too few points to get there.
	private rush(tree: number, talent: number, start: LevelingTalentPick[]): LevelingTalentPick[] | null {
		const shape = this.space.shape;
		const config = shape.trees[tree].talents[talent];
		const route: LevelingTalentPick[] = [];
		const add = (pick: LevelingTalentPick) => {
			if (route.length >= this.routeLength || !this.valid([...route, pick])) return false;
			route.push(pick);
			return true;
		};
		const above = [
			...start.filter(pick => pick.tree == tree),
			...shape.trees[tree].talents
				.map((c, i) => ({ c, i }))
				.sort((a, b) => a.c.row - b.c.row || a.i - b.i)
				.flatMap(({ c, i }) => Array.from({ length: c.maxPoints }, () => LevelingTalentPick.create({ tree, talent: i }))),
		].filter(pick => shape.trees[tree].talents[pick.talent].row < config.row && pick.talent != config.prereq && this.space.allowed(tree, pick.talent));
		const treePoints = () => route.filter(pick => pick.tree == tree).length;
		for (const pick of above) {
			if (treePoints() >= config.row * shape.pointsPerRow) break;
			add(pick);
		}
		if (config.prereq !== undefined) {
			for (let n = 0; n < shape.trees[tree].talents[config.prereq].maxPoints; n++) add(LevelingTalentPick.create({ tree, talent: config.prereq }));
		}
		if (!add(LevelingTalentPick.create({ tree, talent }))) return null;
		for (let n = 1; n < config.maxPoints; n++) add(LevelingTalentPick.create({ tree, talent }));
		return this.extend(route, tree);
	}

	// The best route we find over the levels from the starts. Each start is cut or filled to the
	// points of the last level.
	async search(starts: LevelingTalentPick[][], levels = this.space.levels): Promise<RouteResult> {
		const shape = this.space.shape;
		const rushes = shape.trees.flatMap((tree, t) =>
			tree.talents.flatMap((config, i) => (config.row > 0 && this.space.allowed(t, i) ? [this.rush(t, i, starts[0] ?? [])] : [])),
		);
		const candidates = [...starts.map(start => this.extend(start)), ...shape.trees.map((_, t) => this.extend([], t)), ...rushes].filter(
			(route): route is LevelingTalentPick[] => !!route && route.length == this.routeLength && this.valid(route),
		);
		const unique = [...new Map(candidates.map(route => [JSON.stringify(route), route])).values()];

		// We score the starts at every other level, which is enough to tell the good ones.
		this.progress(`Scoring ${unique.length} starts`, Infinity);
		const startLevels = levels.filter((_, i) => i % 2 == 0 || i == levels.length - 1);
		const scores = await this.score(unique, startLevels);
		const order = unique.map((route, i) => ({ route, hours: scores[i] })).sort((a, b) => a.hours - b.hours);
		const searched = order.slice(0, STARTS_SEARCHED);
		const full = await this.score(
			searched.map(start => start.route),
			levels,
		);
		let best = { route: searched[0].route, hours: full[0] };
		for (const [n, start] of searched.entries()) {
			const found = await this.improve(start.route, full[n], levels, `Start ${n + 1} of ${searched.length}`, best.hours);
			if (found.hours < best.hours) best = found;
			if (this.stopped) break;
		}
		return this.result(best.route, levels);
	}

	private progress(stage: string, bestHours: number) {
		this.onProgress({ stage, sims: this.sims, bestHours });
	}

	// Makes the change that saves the most hours, again and again, until none saves any.
	private async improve(
		route: LevelingTalentPick[],
		hours: number,
		levels: number[],
		stage: string,
		bestSoFar: number,
	): Promise<{ route: LevelingTalentPick[]; hours: number }> {
		let current = { route, hours };
		for (let round = 1; !this.stopped; round++) {
			this.progress(`${stage}, change ${round}`, Math.min(bestSoFar, current.hours));
			const moves = [...this.swaps(current.route, levels), ...(await this.pointMoves(current.route, levels))];
			if (!moves.length || this.stopped) break;
			const scores = await this.score(moves, levels);
			let bestIndex = -1;
			scores.forEach((score, i) => {
				if (score < current.hours * (1 - MIN_GAIN) && (bestIndex < 0 || score < scores[bestIndex])) bestIndex = i;
			});
			if (bestIndex < 0 || this.stopped) break;
			current = { route: moves[bestIndex], hours: scores[bestIndex] };
		}
		return current;
	}

	// Whether some level we score has the pick at index i but not the one at j (i < j), so a
	// swap of the two changes its talents.
	private swapMatters(i: number, j: number, levels: number[]): boolean {
		return levels.some(level => {
			const points = this.space.pointsAt(level);
			return i < points && points <= j;
		});
	}

	private swaps(route: LevelingTalentPick[], levels: number[]): LevelingTalentPick[][] {
		const moves: LevelingTalentPick[][] = [];
		for (let i = 0; i < route.length; i++) {
			for (let j = i + 1; j < Math.min(route.length, i + 1 + SWAP_WINDOW); j++) {
				if (samePick(route[i], route[j]) || !this.swapMatters(i, j, levels)) continue;
				const move = [...route];
				[move[i], move[j]] = [move[j], move[i]];
				if (this.valid(move)) moves.push(move);
			}
		}
		return moves;
	}

	// The routes that take the last point of a talent out and put a point in another talent
	// instead. There are too many to sim each at every level, so we first sim the full build
	// at two levels, the middle and the last, without the talent we take from and with the one
	// we add. The moves that look best by those (taken as if each point counted on its own) get
	// simmed at every level.
	private async pointMoves(route: LevelingTalentPick[], levels: number[]): Promise<LevelingTalentPick[][]> {
		const shape = this.space.shape;
		const counts = countsOf(route, shape);
		const screenLevels = [...new Set([levels[Math.floor((levels.length - 1) / 2)], levels[levels.length - 1]])];

		const moves: Array<{ route: LevelingTalentPick[]; from: LevelingTalentPick; to: LevelingTalentPick }> = [];
		const taken = [...new Map(route.map(pick => [`${pick.tree},${pick.talent}`, pick])).values()];
		for (const from of taken) {
			const index = route.map(pick => samePick(pick, from)).lastIndexOf(true);
			shape.trees.forEach((tree, t) =>
				tree.talents.forEach((config, i) => {
					if (!this.space.allowed(t, i) || (from.tree == t && from.talent == i)) return;
					if (counts[t][i] >= config.maxPoints) return;
					const to = LevelingTalentPick.create({ tree: t, talent: i });
					const inPlace = [...route];
					inPlace[index] = to;
					const atEnd = [...route.slice(0, index), ...route.slice(index + 1), to];
					const move = this.valid(inPlace) ? inPlace : this.valid(atEnd) ? atEnd : null;
					if (move) moves.push({ route: move, from, to });
				}),
			);
		}
		if (!moves.length) return [];

		// A talent's worth at a screening level: the hours there without its last point, or
		// with one more point.
		const withChange = (level: number, pick: LevelingTalentPick, change: number) => {
			const changed = countsOf(route.slice(0, this.space.pointsAt(level)), shape);
			changed[pick.tree][pick.talent] = Math.max(0, changed[pick.tree][pick.talent] + change);
			return countsToTalentsString(changed);
		};
		const jobs: Array<{ level: number; talents: string }> = [];
		for (const level of screenLevels) {
			jobs.push({ level, talents: this.talentsAt(route, level) });
			for (const pick of taken) jobs.push({ level, talents: withChange(level, pick, -1) });
			for (const move of moves) jobs.push({ level, talents: withChange(level, move.to, 1) });
		}
		await this.simMissing(jobs);
		if (this.stopped) return [];
		const screened = moves.map(move => {
			let change = 0;
			for (const level of screenLevels) {
				const now = this.hoursAt(level, this.talentsAt(route, level));
				change += this.hoursAt(level, withChange(level, move.from, -1)) - now + (this.hoursAt(level, withChange(level, move.to, 1)) - now);
			}
			return { move, change };
		});
		screened.sort((a, b) => a.change - b.change);
		return screened.slice(0, SET_MOVES_KEPT).map(s => s.move.route);
	}
}

function samePick(a: LevelingTalentPick, b: LevelingTalentPick): boolean {
	return a.tree == b.tree && a.talent == b.talent;
}
