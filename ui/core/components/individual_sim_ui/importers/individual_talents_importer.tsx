import { IndividualSimUI } from '../../../individual_sim_ui';
import { Spec } from '../../../proto/common';
import { classTalentsConfig } from '../../../talents/factory';
import { TypedEvent } from '../../../typed_event';
import Toast from '../../toast';
import { IndividualImporter } from './individual_importer';

// Takes a bare talent string, the same "digits per tree, dashes between trees" format the
// presets, the sim link and the JSON export carry. A build from a chat message or a
// spreadsheet can then be pasted in without a full settings file around it.
export class IndividualTalentsImporter<SpecType extends Spec> extends IndividualImporter<SpecType> {
	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: 'Talents Import' });

		this.descriptionElem.appendChild(
			<>
				<p>
					Paste a talent string such as <code>005303103014-32003311201513231</code>: one digit per talent in tree order, one block per tree,
					blocks separated by dashes. Trailing zeros can be left off. Only the talents change, gear and settings stay as they are.
				</p>
			</>,
		);
	}

	onImport(data: string) {
		const talentsStr = data.trim();
		if (!/^\d*(-\d*){0,2}$/.test(talentsStr) || talentsStr.replaceAll('-', '') == '') {
			throw new Error('Expected digits separated by dashes, e.g. 005303103014-32003311201513231.');
		}

		const trees = classTalentsConfig[this.simUI.player.getClass()];
		const blocks = talentsStr.split('-');
		if (blocks.length > trees.length) {
			throw new Error(`Found ${blocks.length} trees but this class has ${trees.length}.`);
		}
		let total = 0;
		blocks.forEach((block, treeIdx) => {
			const tree = trees[treeIdx];
			if (block.length > tree.talents.length) {
				throw new Error(`${tree.name} has ${tree.talents.length} talents but the string gives ${block.length}.`);
			}
			[...block].forEach((digit, i) => {
				const points = parseInt(digit);
				const talent = tree.talents[i];
				if (points > talent.maxPoints) {
					throw new Error(`${talent.name || String(talent.fieldName) || `${tree.name} talent ${i + 1}`} has at most ${talent.maxPoints} point${talent.maxPoints == 1 ? '' : 's'}, got ${points}.`);
				}
				total += points;
			});
		});
		if (total > 51) {
			throw new Error(`${total} points spent, the maximum is 51.`);
		}

		// Stored the way the presets spell it, without trailing zeros or empty trees.
		const normalized = blocks
			.map(block => block.replace(/0+$/, ''))
			.join('-')
			.replace(/-+$/, '');
		this.simUI.player.setTalentsString(TypedEvent.nextEventID(), normalized);
		this.close();
		new Toast({ variant: 'success', body: `Talents imported (${total} points).` });
	}
}
