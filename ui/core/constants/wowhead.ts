// Which of wowhead's databases we link into and read tooltips from.
//
// wowhead keeps one database per game and Forever is its own, so a Classic link either
// shows the wrong numbers (Forever reworked a lot of low level gear) or 404s outright on
// anything Forever added, like item 270081. The name goes in the page path
// (wowhead.com/forever/item=...), in the nether tooltip path, and as the dataEnv the
// embedded tooltip script reads.
export const WOWHEAD_DOMAIN = 'forever';
export const WOWHEAD_EXPANSION_ENV = 17;
