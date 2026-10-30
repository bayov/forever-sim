/**
 * API endpoints and exposed wasm function names. Also used as request identifier.
 */
export enum SimRequest {
	bulkSimAsync = 'bulkSimAsync',
	//bulkSimCombos = 'bulkSimCombos',
	computeStats = 'computeStats',
	computeStatsJson = 'computeStatsJson',
	raidSim = 'raidSim',
	raidSimJson = 'raidSimJson',
	raidSimAsync = 'raidSimAsync',
	statWeights = 'statWeights',
	statWeightsAsync = 'statWeightsAsync',
	statWeightRequests = 'statWeightRequests',
	statWeightCompute = 'statWeightCompute',
	raidSimRequestSplit = 'raidSimRequestSplit',
	raidSimResultCombination = 'raidSimResultCombination',
	abortById = 'abortById',
	// The raid sim (docs/raid_sim.md). resolveComp turns a RaidComp into a RaidCompResolution, in the
	// WebAssembly lib and at the native server's /resolveComp. compSearchAsync is the comp search,
	// streamed like raidSimAsync, and only the native server has it.
	resolveComp = 'resolveComp',
	compSearchAsync = 'compSearchAsync',
	// A hand-built raid's member values, at the native server's /compValues (it needs the report's builds).
	compValues = 'compValues',
	// Cancels a request's HTTP call in a net worker (its input is the request id, as UTF-8). The server may
	// go on computing what it was asked; only the page stops waiting.
	abortFetch = 'abortFetch',
}

/**
 * What the Worker receives from the UI
 */
export type WorkerReceiveMessageType = keyof typeof SimRequest | 'setID';

export interface WorkerReceiveMessageBodyBase {
	id: string;
	msg: WorkerReceiveMessageType;
	inputData?: Uint8Array;
}

export interface WorkerReceiveMessageSetId extends WorkerReceiveMessageBodyBase {
	msg: 'setID';
}

export interface WorkerReceiveMessageSimRequest extends Required<WorkerReceiveMessageBodyBase> {
	msg: SimRequest;
}

export type WorkerReceiveMessage = WorkerReceiveMessageSetId | WorkerReceiveMessageSimRequest;

/**
 * What the Worker sends to the UI
 */
export type WorkerSendMessageType = 'ready' | 'idConfirm' | 'progress' | keyof typeof SimRequest;

export interface WorkerSendMessageBodyBase {
	id?: string;
	msg: WorkerSendMessageType;
	outputData?: Uint8Array;
}

export interface WorkerSendMessageIdConfirm extends WorkerSendMessageBodyBase {
	msg: 'idConfirm';
}

export interface WorkerSendMessageReady extends WorkerSendMessageBodyBase {
	msg: 'ready';
}

export interface WorkerSendMessageProgress extends Required<WorkerSendMessageBodyBase> {
	msg: 'progress';
}

export interface WorkerSendMessageSimRequest extends Required<WorkerSendMessageBodyBase>, Required<Omit<WorkerReceiveMessageSimRequest, 'inputData'>> {
	msg: SimRequest;
}

export type WorkerSendMessage = WorkerSendMessageReady | WorkerSendMessageIdConfirm | WorkerSendMessageProgress | WorkerSendMessageSimRequest;
