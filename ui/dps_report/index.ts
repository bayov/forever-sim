/// <reference types="vite/client" />
// The DPS Report page. The same source builds into the sim UI page (/classic/dps_report/) and into a single
// self-contained HTML file for a Claude artifact (make dps-report-artifact).
import tippy from 'tippy.js';

import { installStickyTooltips } from '../core/components/sticky_tooltips';
import { createDash } from './data';
import { loadDataset, manifest, standalone } from './data_source';
import layout from './layout.html?raw';
import { startReport } from './report.js';

tippy.setDefaultProps({ arrow: false, allowHTML: true });

const root = document.getElementById('dps-report')!;
root.innerHTML = layout;
if (standalone) root.querySelector('#home-link')!.removeAttribute('href');
const dash = createDash(manifest, loadDataset, document.getElementById('fv')!);
startReport(dash);
installStickyTooltips();

// Like the sim UI's header, the tabs float over the page once it scrolls. The page scrolls in .sim-ui.
const scroller = root.querySelector('.sim-ui')!;
const header = root.querySelector('.sim-header')!;
scroller.addEventListener('scroll', () => header.classList.toggle('stuck', scroller.scrollTop > 0), { passive: true });
document.body.classList.add('ready');
