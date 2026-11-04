"""md2html.py: the little Markdown the docs use, as simple HTML for the report's text rows (tankheal.py).

Paragraphs, headings (### and deeper become <h4>..<h6>), bullet and numbered lists (nested by indent), bold, italic,
inline code, links and pipe tables. No markdown library is installed here, so this covers what docs/tank_*.md and
docs/healing_*.md write, not all of Markdown. Links to other docs keep their text only (the report can't open them).
"""
import html, re


def inline(t):
    codes = []

    def keep(m):
        codes.append('<code>' + html.escape(m.group(1)) + '</code>')
        return f'\x00{len(codes) - 1}\x00'
    t = re.sub(r'`([^`]+)`', keep, t)
    t = html.escape(t, quote=False)

    def link(m):
        text, url = m.group(1), m.group(2)
        if re.match(r'https?://', url):
            return f'<a href="{html.escape(url)}" target="_blank" rel="noopener">{text}</a>'
        return text  # a repository path: the report has no page for it
    t = re.sub(r'\[([^\]]+)\]\(([^)\s]+)\)', link, t)
    t = re.sub(r'\*\*(.+?)\*\*', r'<b>\1</b>', t)
    t = re.sub(r'(?<![\w*])\*(?!\s)(.+?)(?<!\s)\*(?![\w*])', r'<i>\1</i>', t)
    t = re.sub(r'(?<![\w])_(?!\s)(.+?)(?<!\s)_(?![\w])', r'<i>\1</i>', t)
    return re.sub(r'\x00(\d+)\x00', lambda m: codes[int(m.group(1))], t)


def _cells(line):
    s = line.strip()
    if s.startswith('|'):
        s = s[1:]
    if s.endswith('|'):
        s = s[:-1]
    return [c.strip() for c in s.split('|')]


def convert(md):
    lines = md.replace('\r', '').split('\n')
    out, para, i = [], [], 0
    stack = []  # open lists: (indent, tag)

    def flush():
        if para:
            out.append('<p>' + inline(' '.join(para)) + '</p>')
            para.clear()

    def close_lists(indent=-1):
        while stack and stack[-1][0] > indent:
            out.append(f'</li></{stack.pop()[1]}>')

    while i < len(lines):
        ln = lines[i]
        s = ln.strip()
        if not s:
            flush()
            # a blank line inside a list keeps the list open when the next line continues it
            nxt = next((x for x in lines[i + 1:] if x.strip()), '')
            if stack and not (re.match(r'\s*([-*]|\d+\.)\s', nxt) or (nxt.startswith(' ') and len(nxt) - len(nxt.lstrip()) > stack[0][0])):
                close_lists()
            i += 1
            continue
        if s.startswith('|') and i + 1 < len(lines) and re.match(r'^\s*\|?\s*:?-{3,}', lines[i + 1]):
            flush(), close_lists()
            head = _cells(s)
            rows = []
            i += 2
            while i < len(lines) and lines[i].strip().startswith('|'):
                rows.append(_cells(lines[i]))
                i += 1
            out.append('<table><thead><tr>' + ''.join(f'<th>{inline(c)}</th>' for c in head) + '</tr></thead><tbody>' +
                       ''.join('<tr>' + ''.join(f'<td>{inline(c)}</td>' for c in r) + '</tr>' for r in rows) + '</tbody></table>')
            continue
        m = re.match(r'^(#{1,6})\s+(.*)$', s)
        if m:
            flush(), close_lists()
            lvl = min(6, max(4, len(m.group(1)) + 1))
            out.append(f'<h{lvl}>{inline(m.group(2))}</h{lvl}>')
            i += 1
            continue
        m = re.match(r'^(\s*)([-*]|\d+\.)\s+(.*)$', ln)
        if m:
            flush()
            ind, tag = len(m.group(1)), 'ol' if m.group(2)[0].isdigit() else 'ul'
            if stack and stack[-1][0] == ind and stack[-1][1] != tag:
                close_lists(ind - 1)
            if not stack or ind > stack[-1][0]:
                out.append(f'<{tag}><li>')
                stack.append((ind, tag))
            else:
                close_lists(ind)
                if not stack:
                    out.append(f'<{tag}><li>')
                    stack.append((ind, tag))
                else:
                    out.append('</li><li>')
            out.append(inline(m.group(3)))
            i += 1
            continue
        if stack:  # a list item's continuation line
            out.append(' ' + inline(s))
            i += 1
            continue
        para.append(s)
        i += 1
    flush(), close_lists()
    return ''.join(out)


def sections(md, level=2):
    """[(title, body)] of the doc's sections at this heading level (## by default), the text before the first as ''."""
    out, title, body = [], '', []
    for ln in md.split('\n'):
        m = re.match(r'^(#{%d})\s+(.*)$' % level, ln)
        if m:
            out.append((title, '\n'.join(body)))
            title, body = m.group(2).strip(), []
        else:
            body.append(ln)
    out.append((title, '\n'.join(body)))
    return out
