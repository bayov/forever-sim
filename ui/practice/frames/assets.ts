// The client files the Blizzard frames draw with: textures by FileDataID (the community listfile
// names them), the interface sounds, and the two fonts. Textures load on a frame's first open,
// each file once, and become data URLs the CSS and the inline styles use.

import { border9, canvasOf, fetchFile, glowUrl, image, loadFont, playSound, url } from '../glue';

// Interface textures (interface/...), by the game's path.
export const TX = {
	// buttons/
	panelButtonUp: 130828,
	panelButtonDown: 130825,
	panelButtonHighlight: 130826,
	panelButtonDisabled: 130824,
	closeUp: 130832,
	closeDown: 130830,
	closeHighlight: 130831,
	checkboxUp: 130755,
	checkboxDown: 130752,
	checkboxHighlight: 130753,
	checkboxCheck: 130751,
	sliderBackground: 130858,
	sliderBorder: 130859,
	sliderThumb: 130860,
	scrollUpUp: 130857,
	scrollUpDown: 130855,
	scrollUpDisabled: 130854,
	scrollUpHighlight: 130856,
	scrollDownUp: 130853,
	scrollDownDown: 130851,
	scrollDownDisabled: 130850,
	scrollDownHighlight: 130852,
	scrollKnob: 130849,
	quickslot2: 130841,
	quickslotDepress: 130839,
	buttonHilightSquare: 130718,
	checkButtonHilight: 130724,
	emptySlot: 130766,
	emptySlotWhite: 130765,
	mouseHilight: 130757,
	prevPageUp: 130869,
	prevPageDown: 130868,
	prevPageDisabled: 130867,
	nextPageUp: 130866,
	nextPageDown: 130865,
	nextPageDisabled: 130864,
	// dialogframe/
	dialogBorder: 131072,
	dialogBackground: 131071,
	dialogBackgroundDark: 312922,
	dialogHeader: 131080,
	// common/
	inputBorder: 130975,
	// paperdollinfoframe/
	characterL1: 136548,
	characterR1: 136550,
	characterL2: 136549,
	characterR2: 136551,
	characterBottomLeft: 136546,
	characterBottomRight: 136547,
	generalTopLeft: 136558,
	generalTopRight: 136559,
	generalBottomLeft: 136556,
	generalBottomRight: 136557,
	skillsTopLeft: 136552,
	skillsTopRight: 136554,
	skillsL2: 136553,
	skillsR2: 136555,
	skillFrameBotLeft: 136536,
	skillFrameBotRight: 136537,
	activeTab: 136544,
	inactiveTab: 136564,
	tabHighlight: 136580,
	statBackground: 136579,
	resistanceIcons: 136568,
	skillsBar: 136570,
	skillsBarBorder: 136571,
	skillsBarBorderHighlight: 136572,
	reputationDetailBackground: 136565,
	honorTopLeft: 136562,
	honorTopRight: 136563,
	honorBottomLeft: 136560,
	honorBottomRight: 136561,
	ammoSlot: 136545,
	// paperdoll/
	slotHead: 136516,
	slotNeck: 136519,
	slotShoulder: 136526,
	slotBack: 136521,
	slotChest: 136512,
	slotShirt: 136525,
	slotTabard: 136527,
	slotWrists: 136530,
	slotHands: 136515,
	slotWaist: 136529,
	slotLegs: 136517,
	slotFeet: 136513,
	slotFinger: 136514,
	slotTrinket: 136528,
	slotMainHand: 136518,
	slotSecondaryHand: 136524,
	slotRanged: 136520,
	slotRelic: 136522,
	slotAmmo: 136510,
	slotBag: 136511,
	backpackEmptySlot: 136509,
	// talentframe/
	talentBotLeft: 136963,
	talentBotRight: 136964,
	talentRankBorder: 136960,
	talentArrows: 136961,
	talentBranches: 136962,
	// spellbook/
	spellbookIcon: 136830,
	spellbookSkillLineTab: 136831,
	spellbookSpellBackground: 136832,
	spellbookTabUnselected: 136833,
	spellbookTab1Selected: 136834,
	spellbookTab3Selected: 136835,
	spellbookBotLeft: 136836,
	spellbookBotRight: 136837,
	spellbookTabHighlight: 136838,
	spellbookTopLeft: 136839,
	spellbookTopRight: 136840,
	// containerframe/
	backpackBackground: 130981,
	bag1Slot: 130982,
	bagComponents: 131003,
	bagMiddle: 131005,
	bag4x4: 130998,
	// optionsframe/
	optionsTopLeft: 136505,
	optionsTopRight: 136506,
	optionsBottomLeft: 136503,
	optionsBottomRight: 136504,
	optionsActiveTab: 136500,
	optionsInactiveTab: 136501,
	// chatframe/
	chatBackground: 130937,
	chatTab: 130942,
	chatTabBgLeft: 374162,
	chatTabBgMid: 374164,
	chatInputLeft: 130959,
	chatInputRight: 130960,
	chatInput: 130961,
	chatScrollUp: 130958,
	chatScrollDown: 130952,
	chatScrollEnd: 130955,
	chatTabBgRight: 374166,
	chatTabSelectedLeft: 374174,
	chatTabSelectedMid: 374176,
	chatTabSelectedRight: 374178,
	chatTabHighlightLeft: 374168,
	chatTabHighlightMid: 374170,
	chatTabHighlightRight: 374172,
	// common/, targetingframe/ and tooltips/
	whiteIconFrame: 651080,
	classCircles: 237669,
	tooltipBorder: 137057,
};

// The talent trees' backgrounds (talentframe/<name>-topleft, then topright, bottomleft and
// bottomright at the next ids), by the client's background file name.
export const talentBackgrounds: Record<string, number> = {
	DruidBalance: 136878,
	DruidFeralCombat: 136882,
	DruidRestoration: 136886,
	HunterBeastMastery: 136890,
	HunterMarksmanship: 136894,
	HunterSurvival: 136898,
	MageArcane: 136902,
	MageFire: 136906,
	MageFrost: 136910,
	PaladinCombat: 136914,
	PaladinHoly: 136918,
	PaladinProtection: 136922,
	PriestDiscipline: 136926,
	PriestHoly: 136930,
	PriestShadow: 136934,
	RogueAssassination: 136938,
	RogueCombat: 136942,
	RogueSubtlety: 136946,
	ShamanElementalCombat: 136950,
	ShamanEnhancement: 136954,
	ShamanRestoration: 136958,
	WarlockCurses: 136967,
	WarlockDestruction: 136971,
	WarlockSummoning: 136975,
	WarriorArms: 136983,
	WarriorFury: 136987,
	WarriorProtection: 136991,
};

// The interface sounds (sound/interface), by the game's sound kit names.
export const SND = {
	characterOpen: 567507, // igCharacterInfoOpen
	characterClose: 567433, // igCharacterInfoClose
	characterTab: 567422, // igCharacterInfoTab
	spellbookOpen: 567440, // igSpellBookOpen
	spellbookClose: 567496, // igSpellBookClose
	pageTurn: 567472, // igAbiliityPageTurn
	talentOpen: 567412, // TalentScreenOpen
	talentClose: 567424, // TalentScreenClose
	bagOpen: 567543, // igBackPackOpen
	bagClose: 567559, // igBackPackClose
	menuOpen: 567490, // igMainMenuOpen
	menuClose: 567464, // igMainMenuClose
	menuOption: 567481, // igMainMenuOption
	checkboxOn: 567481, // igMainMenuOptionCheckBoxOn
	chatScroll: 567407, // uChatScrollButton
};

export function sound(fdid: number) {
	playSound(fdid);
}

// Fonts: Friz Quadrata (glue.ts loads it) for the frames, Arial Narrow for chat and numbers.
const arialNarrow = 615958;
let fontsReady: Promise<void> | undefined;
export function loadFonts(): Promise<void> {
	fontsReady ??= Promise.all([
		loadFont(),
		fetchFile(arialNarrow)
			.then(buf => new FontFace('Arial Narrow WoW', buf).load())
			.then(f => void document.fonts.add(f))
			.catch(e => console.warn('chat font', e)),
	]).then(() => undefined);
	return fontsReady;
}

const urls = new Map<string, Promise<string>>();

function cached(key: string, make: () => Promise<string>): Promise<string> {
	let p = urls.get(key);
	if (!p) {
		p = make();
		p.catch(() => urls.delete(key));
		urls.set(key, p);
	}
	return p;
}

// tex gives a texture as a data URL. A region (in the file's pixels) cuts a part out of it.
export function tex(fdid: number, region?: [x: number, y: number, w: number, h: number]): Promise<string> {
	if (!region) return cached(String(fdid), () => url(fdid));
	return cached(`${fdid}:${region.join(',')}`, async () => canvasOf(await image(fdid), ...region).toDataURL());
}

// glow is a texture the game draws with additive blending (highlights, glows), with its
// brightness turned into alpha so it looks the same drawn normally.
export function glow(fdid: number): Promise<string> {
	return cached(`glow:${fdid}`, () => glowUrl(fdid));
}

// texSize gives a texture's size in pixels.
export async function texSize(fdid: number): Promise<{ width: number; height: number }> {
	const img = await image(fdid);
	return { width: img.width, height: img.height };
}

const vars = new Map<string, Promise<void>>();

// texVars sets textures as CSS variables on the page (each once), for the CSS of shared parts
// like buttons and checkboxes.
export function texVars(list: Record<string, () => Promise<string>>): Promise<void> {
	return Promise.all(
		Object.entries(list).map(([name, make]) => {
			let p = vars.get(name);
			if (!p) {
				p = make().then(u => document.documentElement.style.setProperty(`--${name}`, `url(${u})`));
				p.catch(() => vars.delete(name));
				vars.set(name, p);
			}
			return p;
		}),
	).then(() => undefined);
}

// The shared parts' textures: panel buttons, the close button, tabs, checkboxes, sliders, scroll
// bars and the dialog border. Every frame waits for these on its first open.
export function sharedVars(): Promise<void> {
	return texVars({
		'bz-btn-up': () => tex(TX.panelButtonUp, [0, 0, 80, 22]),
		'bz-btn-down': () => tex(TX.panelButtonDown, [0, 0, 80, 22]),
		'bz-btn-disabled': () => tex(TX.panelButtonDisabled, [0, 0, 80, 22]),
		'bz-btn-highlight': () => glow(TX.panelButtonHighlight).then(u => crop(u, 0, 0, 80, 22)),
		'bz-close-up': () => tex(TX.closeUp),
		'bz-close-down': () => tex(TX.closeDown),
		'bz-close-highlight': () => glow(TX.closeHighlight),
		'bz-tab-active': () => tex(TX.activeTab),
		'bz-tab-inactive': () => tex(TX.inactiveTab),
		'bz-tab-highlight': () => glow(TX.tabHighlight),
		'bz-check-up': () => tex(TX.checkboxUp),
		'bz-check-down': () => tex(TX.checkboxDown),
		'bz-check-highlight': () => glow(TX.checkboxHighlight),
		'bz-check-check': () => tex(TX.checkboxCheck),
		'bz-slider-bg': () => tex(TX.sliderBackground),
		'bz-slider-border': () => cached('slider9', () => border9(TX.sliderBorder)),
		'bz-slider-thumb': () => tex(TX.sliderThumb),
		'bz-scroll-up': () => tex(TX.scrollUpUp),
		'bz-scroll-up-down': () => tex(TX.scrollUpDown),
		'bz-scroll-up-disabled': () => tex(TX.scrollUpDisabled),
		'bz-scroll-down': () => tex(TX.scrollDownUp),
		'bz-scroll-down-down': () => tex(TX.scrollDownDown),
		'bz-scroll-down-disabled': () => tex(TX.scrollDownDisabled),
		'bz-scroll-highlight': () => glow(TX.scrollUpHighlight),
		'bz-scroll-knob': () => tex(TX.scrollKnob),
		'bz-dialog-bg': () => tex(TX.dialogBackground),
		'bz-dialog-bg-dark': () => tex(TX.dialogBackgroundDark),
		'bz-dialog-border': () => cached('border9', () => border9(TX.dialogBorder)),
		'bz-dialog-header': () => tex(TX.dialogHeader),
		'bz-input-border': () => tex(TX.inputBorder),
		'bz-slot': () => tex(TX.quickslot2),
		'bz-slot-highlight': () => glow(TX.buttonHilightSquare),
		'bz-slot-pushed': () => tex(TX.quickslotDepress),
		'bz-icon-border': () => tex(TX.whiteIconFrame),
	});
}

async function crop(dataUrl: string, x: number, y: number, w: number, h: number): Promise<string> {
	const img = new Image();
	img.src = dataUrl;
	await img.decode();
	const c = document.createElement('canvas');
	c.width = w;
	c.height = h;
	c.getContext('2d')!.drawImage(img, x, y, w, h, 0, 0, w, h);
	return c.toDataURL();
}
