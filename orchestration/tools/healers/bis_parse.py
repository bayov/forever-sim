import re, sys, json
def picks(path):
    h = open(path).read()
    h = h.replace('\\/', '/').replace('\\n', '\n').replace('\\"', '"')
    out = []
    for m in re.finditer(r'\[h3 toc="([^"]+)"\](.*?)(?=\[h3 toc=|\[h2 |$)', h, re.S):
        slot, body = m.group(1), m.group(2)
        rows = re.findall(r'\[tr\](.*?)\[/tr\]', body, re.S)
        for r in rows:
            cells = re.findall(r'\[td[^\]]*\](.*?)\[/td\]', r, re.S)
            if not cells:
                continue
            rank = re.sub(r'\[.*?\]', '', cells[0]).strip()
            items = re.findall(r'\[item=(\d+)\]([^\[\n]*)', cells[1]) if len(cells) > 1 else []
            if items:
                out.append((slot, rank, [(int(i), t.strip()) for i, t in items]))
    return out
for p in sys.argv[1:]:
    print('###', p)
    for slot, rank, items in picks(p):
        print(f'{slot:<22} {rank:<10} {items}')
