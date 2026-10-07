import clsx from 'clsx';

import { getLaunchedSimsForClass, LaunchStatus, raidSimStatus, simLaunchStatuses } from '../launched_sims.js';
import { Class, Spec } from '../proto/common.js';
import {
	classIcons,
	classNames,
	homeSiteUrl,
	raidSimIcon,
	raidSimLabel,
	specNames,
	specToClass,
	textCssClassForClass,
	textCssClassForSpec,
	titleIcons,
} from '../proto_utils/utils.js';
import { Component } from './component.js';

interface ClassOptions {
	type: 'Class';
	index: Class;
}

interface SpecOptions {
	type: 'Spec';
	index: Spec;
}

interface RaidOptions {
	type: 'Raid';
}

type SimTitleDropdownConfig = {
	noDropdown?: boolean;
};

// The sim's title at the top of the sidebar. Clicking it takes us back to the landing page,
// where every spec is one click away. In the raid sim's player editor it's only a label.
export class SimTitleDropdown extends Component {
	constructor(parent: HTMLElement, currentSpecIndex: Spec | null, config: SimTitleDropdownConfig = {}) {
		super(parent, 'sim-title-dropdown-root');

		const rootLinkArgs: SpecOptions | RaidOptions = currentSpecIndex === null ? { type: 'Raid' } : { type: 'Spec', index: currentSpecIndex };
		this.rootElem.appendChild(this.buildRootSimLink(rootLinkArgs, config.noDropdown ? undefined : homeSiteUrl));
	}

	// When the title is a link, it also holds a second text, which takes the place of the
	// sim's name when we point at it and says where the link goes.
	private buildRootSimLink(data: SpecOptions | RaidOptions, href?: string): Element {
		let label;

		if (data.type == 'Raid') label = raidSimLabel;
		else {
			const classIndex = specToClass[data.index];
			if (getLaunchedSimsForClass(classIndex).length > 1)
				// If the class has multiple sims, use the spec name
				label = specNames[data.index];
			// If the class has only 1 sim, use the class name
			else label = classNames[classIndex];
		}

		return (
			<a href={href ?? 'javascript:void(0)'} className={clsx('sim-link', href && 'sim-title-home-link', this.getContextualKlass(data))}>
				<div className="sim-link-content">
					<img src={this.getSimIconPath(data)} className="sim-link-icon" />
					<div className="sim-title-texts">
						<div className="sim-title-current d-flex flex-column">
							<span className="sim-link-label text-white">WoWSims - Forever</span>
							<span className="sim-link-title">{label}</span>
							{this.launchStatusLabel(data)}
						</div>
						{href && (
							<div className="sim-title-back d-flex flex-column">
								<span className="sim-title-back-title">
									<i className="fas fa-arrow-left sim-title-back-arrow" />
									Change spec
								</span>
								<span className="sim-link-label">Back to class & spec list</span>
							</div>
						)}
					</div>
				</div>
			</a>
		);
	}

	private launchStatusLabel(data: SpecOptions | RaidOptions): Element {
		const status = data.type == 'Raid' ? raidSimStatus.status : simLaunchStatuses[data.index].status;
		const phase = data.type == 'Raid' ? raidSimStatus.phase : simLaunchStatuses[data.index].phase;

		return (
			<span className="launch-status-label text-brand">
				{status === LaunchStatus.Unlaunched ? (
					<>Not Yet Supported</>
				) : (
					<>
						Phase {phase}
						{status != LaunchStatus.Launched && <> - {LaunchStatus[status]}</>}
					</>
				)}
			</span>
		);
	}

	private getSimIconPath(data: ClassOptions | SpecOptions | RaidOptions): string {
		let iconPath: string;

		if (data.type == 'Raid') {
			iconPath = raidSimIcon;
		} else if (data.type == 'Class') {
			iconPath = classIcons[data.index];
		} else {
			iconPath = titleIcons[data.index];
		}

		return iconPath;
	}

	private getContextualKlass(data: ClassOptions | SpecOptions | RaidOptions): string {
		if (data.type == 'Raid')
			// Raid link
			return 'text-white';
		else if (data.type == 'Class')
			// Class links
			return textCssClassForClass(data.index);
		else return textCssClassForSpec(data.index);
	}
}
