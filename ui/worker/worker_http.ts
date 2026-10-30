import { HandlerFunction, WorkerInterface } from './worker_interface';

const defaultRequestOptions = {
	method: 'POST',
	headers: {
		'Content-Type': 'application/x-protobuf',
	},
};

const concatBytes = (a: Uint8Array, b: Uint8Array): Uint8Array => {
	if (a.length === 0) return b;
	const out = new Uint8Array(a.length + b.length);
	out.set(a);
	out.set(b, a.length);
	return out;
};

export const setupHttpWorker = (baseURL: string) => {
	// The HTTP calls abortFetch can cancel, by request id.
	const controllers = new Map<string, AbortController>();

	const makeHttpApiRequest = (endPoint: string, inputData: Uint8Array, requestId: string, signal?: AbortSignal) =>
		fetch(`${baseURL}/${endPoint}?requestId=${requestId}`, {
			...defaultRequestOptions,
			body: inputData,
			signal,
		});

	const syncHandler: HandlerFunction = async (inputData, _, id, msg) => {
		const response = await makeHttpApiRequest(msg, inputData, id);
		const ab = await response.arrayBuffer();
		return new Uint8Array(ab);
	};

	// The server streams progress as frames of a 4 byte big endian length and a
	// ProgressMetrics, and the last frame holds the final result.
	const asyncHandler: HandlerFunction = async (inputData, progress, id, msg) => {
		const response = await makeHttpApiRequest(msg, inputData, id);
		if (!response.ok || !response.body) {
			console.error(`${msg} failed with HTTP ${response.status}`);
			return new Uint8Array();
		}

		const reader = response.body.getReader();
		let buffer = new Uint8Array();
		let outputData = new Uint8Array();
		while (true) {
			const { done, value } = await reader.read();
			if (done) break;
			buffer = concatBytes(buffer, value);

			while (buffer.length >= 4) {
				const frameLength = new DataView(buffer.buffer, buffer.byteOffset, 4).getUint32(0);
				if (buffer.length < 4 + frameLength) break;
				outputData = buffer.slice(4, 4 + frameLength);
				buffer = buffer.slice(4 + frameLength);
				progress(outputData);
			}
		}
		return outputData;
	};

	// A server built before an endpoint answers it with its file server's 404 page, which isn't a
	// proto. So a call that fails answers no bytes, and the caller reads that as "not in this build".
	// abortFetch cancels it, which also answers no bytes.
	const syncHandlerOrEmpty: HandlerFunction = async (inputData, _, id, msg) => {
		const controller = new AbortController();
		controllers.set(id, controller);
		try {
			const response = await makeHttpApiRequest(msg, inputData, id, controller.signal);
			if (!response.ok) {
				console.error(`${msg} failed with HTTP ${response.status}`);
				return new Uint8Array();
			}
			return new Uint8Array(await response.arrayBuffer());
		} catch (e) {
			if (controller.signal.aborted) return new Uint8Array();
			throw e;
		} finally {
			controllers.delete(id);
		}
	};

	const abortFetch: HandlerFunction = inputData => {
		const id = new TextDecoder().decode(inputData);
		controllers.get(id)?.abort();
		return new Uint8Array([controllers.has(id) ? 1 : 0]);
	};

	const noWasmConcurrency: HandlerFunction = (inputData, progress, msg) => {
		const errmsg = `Tried to use ${msg} while using a http worker! This is only supported for wasm!`;
		console.error(errmsg);
		return new Uint8Array();
	};

	new WorkerInterface({
		bulkSimAsync: asyncHandler,
		//bulkSimCombos: syncHandler,
		computeStats: syncHandler,
		computeStatsJson: syncHandler,
		raidSim: syncHandler,
		raidSimJson: syncHandler,
		raidSimAsync: asyncHandler,
		statWeights: syncHandler,
		statWeightsAsync: asyncHandler,
		statWeightRequests: syncHandler,
		statWeightCompute: syncHandler,
		raidSimRequestSplit: noWasmConcurrency,
		raidSimResultCombination: noWasmConcurrency,
		abortById: syncHandler,
		resolveComp: syncHandlerOrEmpty,
		compSearchAsync: asyncHandler,
		compValues: syncHandlerOrEmpty,
		abortFetch,
	}).ready(false);
};
