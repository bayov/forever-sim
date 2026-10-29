import { WorkerInterface } from './worker_interface';

type SimRequestAsync = (data: Uint8Array, progress: (result: Uint8Array) => void, id: string) => Uint8Array;
type SimRequestSync = (data: Uint8Array) => Uint8Array;

// Functions provided or used by the wasm lib.
declare global {
	function wasmready(): void;
	const bulkSimAsync: SimRequestAsync;
	const bulkSimCombos: SimRequestSync;
	const computeStats: SimRequestSync;
	const computeStatsJson: SimRequestSync;
	const raidSim: SimRequestSync;
	const raidSimJson: SimRequestSync;
	const raidSimAsync: SimRequestAsync;
	const statWeights: SimRequestSync;
	const statWeightsAsync: SimRequestAsync;
	const statWeightRequests: SimRequestSync;
	const statWeightCompute: SimRequestSync;
	const raidSimResultCombination: SimRequestSync;
	const raidSimRequestSplit: SimRequestSync;
	const abortById: SimRequestSync;
	// Missing from a lib built before the raid sim's resolver.
	const resolveComp: SimRequestSync | undefined;
}

// What a lib without a function answers: no bytes. The page that called it reads that as "not in this
// build" (see WorkerPool.resolveComp). The comp search is never in the lib, since it needs the native
// server's cache and the report's requests on disk.
const missing = (name: string) => () => {
	console.error(`${name} isn't in this build of the sim's WebAssembly lib.`);
	return new Uint8Array();
};

// Wasm binary calls this function when its done loading.
// eslint-disable-next-line @typescript-eslint/no-unused-vars
globalThis.wasmready = function () {
	new WorkerInterface({
		bulkSimAsync: bulkSimAsync,
		//bulkSimCombos: bulkSimCombos,
		computeStats: computeStats,
		computeStatsJson: computeStatsJson,
		raidSim: raidSim,
		raidSimJson: raidSimJson,
		raidSimAsync: raidSimAsync,
		statWeights: statWeights,
		statWeightsAsync: statWeightsAsync,
		statWeightRequests: statWeightRequests,
		statWeightCompute: statWeightCompute,
		raidSimRequestSplit: raidSimRequestSplit,
		raidSimResultCombination: raidSimResultCombination,
		abortById: abortById,
		resolveComp: typeof resolveComp === 'function' ? resolveComp : missing('resolveComp'),
		compSearchAsync: missing('compSearchAsync'),
		compValues: missing('compValues'),
	}).ready(true);
};

const go = new Go();
let inst: WebAssembly.Instance | null = null;

WebAssembly.instantiateStreaming(fetch('lib.wasm'), go.importObject).then(async result => {
	inst = result.instance;
	// console.log("loading wasm...")
	await go.run(inst);
});

export {};
