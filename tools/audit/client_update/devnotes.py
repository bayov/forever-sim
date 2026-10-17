"""usage: python3 -I devnotes.py <topic.json> [all]
Prints a Blizzard forum thread's posts as text, blue posters only unless 'all' is given, with
created and updated times (an edited post keeps its date; compare 'upd' to catch quiet edits).
Fetch the JSON first, slowly (the forum rate-limits), with the slug URL:
  curl -sL --max-time 30 'https://us.forums.blizzard.com/en/wow/t/<slug>/<topicID>.json?print=true' -o topic.json
The development notes thread is topic 2360696; its slug changes with each update, so start from
https://us.forums.blizzard.com/en/wow/t/2360696.json?print=true and follow the redirect."""
import html, json, re, sys
BLUE = ('Kaivax', 'Aggrend', 'Gnedi', 'Linxy', 'Brenoch')
d = json.load(open(sys.argv[1]))
print('TITLE', d.get('title'), d.get('posts_count'))
for p in d['post_stream']['posts']:
    if p['username'] not in BLUE and len(sys.argv) < 3: continue
    t = p['cooked']
    t = re.sub(r'<li>', '\n- ', t); t = re.sub(r'<h\d[^>]*>', '\n## ', t); t = re.sub(r'</?p>', '\n', t); t = re.sub(r'<br\s*/?>', '\n', t)
    t = re.sub(r'<[^>]+>', '', t)
    print('=====', p['post_number'], p['username'], p['created_at'][:16], 'upd', p.get('updated_at', '')[:16])
    print(re.sub(r'\n\s*\n+', '\n', html.unescape(t)))
