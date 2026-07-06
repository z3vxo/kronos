#!/usr/bin/env python3
import json
import os
import glob
from datetime import datetime

KRONOS_DIR = os.path.expanduser("~/.kronos/logs")
OPERATORS_LOG = os.path.join(KRONOS_DIR, "operators.jsonl")
TASKS_DIR = os.path.join(KRONOS_DIR, "tasks")
OUTPUT_FILE = os.path.expanduser("~/timeline.html")

events = []

if os.path.exists(OPERATORS_LOG):
    with open(OPERATORS_LOG) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                e = json.loads(line)
                ts = e.get("time", "")
                if isinstance(ts, str):
                    try:
                        dt = datetime.fromisoformat(ts)
                        unix = dt.timestamp()
                    except:
                        unix = 0
                else:
                    unix = float(ts)
                events.append({
                    "time": unix,
                    "type": "operator",
                    "event": e.get("event", "unknown"),
                    "msg": e.get("msg", ""),
                    "details": {k: v for k, v in e.items() if k not in ("time", "level", "msg", "event")},
                })
            except json.JSONDecodeError:
                continue

if os.path.exists(TASKS_DIR):
    for guid_dir in glob.glob(os.path.join(TASKS_DIR, "*")):
        if not os.path.isdir(guid_dir):
            continue
        guid = os.path.basename(guid_dir)
        meta_path = os.path.join(guid_dir, "meta.jsonl")
        if not os.path.exists(meta_path):
            continue
        with open(meta_path) as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    e = json.loads(line)
                    finished = e.get("finished_at", 0)
                    tasked = e.get("tasked_at", 0)
                    task_id = e.get("task_id", "?")
                    cmd_name = e.get("cmd_name", f"cmd_{e.get('cmd_code', '?')}")
                    param1 = e.get("param_1", "")
                    param2 = e.get("param_2", "")

                    if tasked:
                        events.append({
                            "time": float(tasked),
                            "type": "task_issued",
                            "guid": guid,
                            "task_id": task_id,
                            "cmd": cmd_name,
                            "param1": param1,
                            "param2": param2,
                        })

                    if finished:
                        output_path = os.path.join(guid_dir, "output", f"{task_id}.txt")
                        output_preview = ""
                        if os.path.exists(output_path):
                            with open(output_path, errors="replace") as of:
                                output_preview = of.read(2000)

                        events.append({
                            "time": float(finished),
                            "type": "task_complete",
                            "guid": guid,
                            "task_id": task_id,
                            "cmd": cmd_name,
                            "output_preview": output_preview,
                        })
                except json.JSONDecodeError:
                    continue

events.sort(key=lambda x: x["time"])

def fmt_time(unix):
    if not unix:
        return "?"
    return datetime.fromtimestamp(unix).strftime("%H:%M:%S")

def fmt_date(unix):
    if not unix:
        return "?"
    return datetime.fromtimestamp(unix).strftime("%Y-%m-%d")

def escape(s):
    return str(s).replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;").replace('"', "&quot;")

# Group events by date
days = {}
for e in events:
    d = fmt_date(e["time"])
    days.setdefault(d, []).append(e)

# Count stats
n_reg = sum(1 for e in events if e["type"] == "operator" and "register" in e.get("event", ""))
n_issued = sum(1 for e in events if e["type"] == "task_issued")
n_complete = sum(1 for e in events if e["type"] == "task_complete")
n_agents = len(set(e.get("guid", "") for e in events if e.get("guid")))
n_logins = sum(1 for e in events if e["type"] == "operator" and e.get("event") == "login")

# Build event rows
def render_event(e):
    ts = fmt_time(e["time"])
    t = e["type"]

    if t == "operator":
        evt = e["event"]
        details = {k: v for k, v in e["details"].items()}
        detail_str = " ".join(f"{k}={v}" for k, v in details.items())

        if "register" in evt:
            codename = details.get("codename", "?")
            host = details.get("hostname", "?")
            user = details.get("user", "?")
            ip = details.get("ip", "?")
            guid = details.get("guid", "?")
            return f'''<div class="ev ev--register">
                <div class="ev__time">{ts}</div>
                <div class="ev__body">
                    <span class="ev__tag ev__tag--register">REGISTER</span>
                    <span class="ev__codename">{escape(codename)}</span>
                    <span class="ev__meta">{escape(user)}@{escape(host)}</span>
                    <span class="ev__meta ev__meta--dim">{escape(ip)}</span>
                    <span class="ev__meta ev__meta--dim">guid={escape(str(guid))}</span>
                </div>
            </div>'''
        elif "login" in evt:
            user = details.get("user", "?")
            return f'''<div class="ev ev--login">
                <div class="ev__time">{ts}</div>
                <div class="ev__body">
                    <span class="ev__tag ev__tag--login">LOGIN</span>
                    <span class="ev__meta">{escape(user)}</span>
                </div>
            </div>'''
        elif "create" in evt:
            proto = details.get("proto", "?")
            port = details.get("port", "?")
            operator = details.get("operator", "?")
            return f'''<div class="ev ev--infra">
                <div class="ev__time">{ts}</div>
                <div class="ev__body">
                    <span class="ev__tag ev__tag--infra">LISTENER+</span>
                    <span class="ev__meta">{escape(proto)}://:{escape(str(port))}</span>
                    <span class="ev__meta ev__meta--dim">by {escape(operator)}</span>
                </div>
            </div>'''
        elif "stop" in evt:
            lid = details.get("id", "?")[:8]
            operator = details.get("operator", "?")
            return f'''<div class="ev ev--infra">
                <div class="ev__time">{ts}</div>
                <div class="ev__body">
                    <span class="ev__tag ev__tag--stop">LISTENER-</span>
                    <span class="ev__meta">{escape(lid)}</span>
                    <span class="ev__meta ev__meta--dim">by {escape(operator)}</span>
                </div>
            </div>'''
        elif "restore" in evt:
            proto = details.get("protocol", "?")
            port = details.get("port", "?")
            return f'''<div class="ev ev--infra">
                <div class="ev__time">{ts}</div>
                <div class="ev__body">
                    <span class="ev__tag ev__tag--restore">RESTORE</span>
                    <span class="ev__meta">{escape(proto)}://:{escape(str(port))}</span>
                </div>
            </div>'''
        else:
            return f'''<div class="ev ev--infra">
                <div class="ev__time">{ts}</div>
                <div class="ev__body">
                    <span class="ev__tag ev__tag--infra">{escape(evt).upper()}</span>
                    <span class="ev__meta">{escape(detail_str)}</span>
                </div>
            </div>'''

    elif t == "task_issued":
        params = e.get("param1", "")
        if e.get("param2"):
            params += f" {e['param2']}"
        cmd = e.get("cmd", "?")
        return f'''<div class="ev ev--issued">
            <div class="ev__time">{ts}</div>
            <div class="ev__body">
                <span class="ev__tag ev__tag--issued">TASK</span>
                <span class="ev__cmd">{escape(cmd)}</span>
                <span class="ev__params">{escape(params)}</span>
                <span class="ev__meta ev__meta--dim">tid={e["task_id"]}</span>
            </div>
        </div>'''

    elif t == "task_complete":
        cmd = e.get("cmd", "?")
        preview = e.get("output_preview", "").strip()
        output_html = ""
        if preview:
            output_html = f'''<details class="ev__output-wrap">
                <summary class="ev__output-toggle">{len(preview)} bytes</summary>
                <pre class="ev__output">{escape(preview)}</pre>
            </details>'''
        return f'''<div class="ev ev--complete">
            <div class="ev__time">{ts}</div>
            <div class="ev__body">
                <span class="ev__tag ev__tag--complete">RECV</span>
                <span class="ev__cmd">{escape(cmd)}</span>
                <span class="ev__meta ev__meta--dim">tid={e["task_id"]}</span>
                {output_html}
            </div>
        </div>'''

    return ""

all_rows = []
for day, day_events in days.items():
    all_rows.append(f'<div class="day-mark">{day}</div>')
    for e in day_events:
        all_rows.append(render_event(e))

html = f'''<title>Kronos Timeline</title>
<style>
:root {{
    --bg: #0c1017;
    --bg-raised: #111820;
    --bg-output: #080c12;
    --fg: #8a919a;
    --fg-bright: #c8cdd3;
    --fg-dim: #4a5260;
    --accent: #c49a3c;
    --accent-dim: #8a6e2e;
    --green: #5a8f5c;
    --violet: #7b6b99;
    --teal: #4a7a6a;
    --steel: #6a7a8a;
    --red: #8f5a5a;
    --amber: #9a7a3c;
    --rule: #1a2030;
    --mono: 'SF Mono', 'Cascadia Code', 'JetBrains Mono', 'Fira Code', Consolas, monospace;
}}

@media (prefers-color-scheme: light) {{
    :root {{
        --bg: #f4f5f7;
        --bg-raised: #ffffff;
        --bg-output: #ebedf0;
        --fg: #4a5060;
        --fg-bright: #1a1e28;
        --fg-dim: #9aa0aa;
        --accent: #9a7520;
        --accent-dim: #b8a060;
        --green: #3a6a3c;
        --violet: #5a4a7a;
        --teal: #3a6a5a;
        --steel: #5a6a7a;
        --red: #7a4040;
        --amber: #7a6030;
        --rule: #dde0e6;
    }}
}}
:root[data-theme="light"] {{
    --bg: #f4f5f7;
    --bg-raised: #ffffff;
    --bg-output: #ebedf0;
    --fg: #4a5060;
    --fg-bright: #1a1e28;
    --fg-dim: #9aa0aa;
    --accent: #9a7520;
    --accent-dim: #b8a060;
    --green: #3a6a3c;
    --violet: #5a4a7a;
    --teal: #3a6a5a;
    --steel: #5a6a7a;
    --red: #7a4040;
    --amber: #7a6030;
    --rule: #dde0e6;
}}
:root[data-theme="dark"] {{
    --bg: #0c1017;
    --bg-raised: #111820;
    --bg-output: #080c12;
    --fg: #8a919a;
    --fg-bright: #c8cdd3;
    --fg-dim: #4a5260;
    --accent: #c49a3c;
    --accent-dim: #8a6e2e;
    --green: #5a8f5c;
    --violet: #7b6b99;
    --teal: #4a7a6a;
    --steel: #6a7a8a;
    --red: #8f5a5a;
    --amber: #9a7a3c;
    --rule: #1a2030;
}}

* {{ margin: 0; padding: 0; box-sizing: border-box; }}
body {{
    background: var(--bg);
    color: var(--fg);
    font-family: var(--mono);
    font-size: 13px;
    line-height: 1.5;
    font-variant-numeric: tabular-nums;
    -webkit-font-smoothing: antialiased;
}}

.shell {{
    max-width: 860px;
    margin: 0 auto;
    padding: 40px 24px 80px;
}}

.header {{
    margin-bottom: 32px;
    border-bottom: 1px solid var(--rule);
    padding-bottom: 20px;
}}

.header__title {{
    font-size: 14px;
    font-weight: 600;
    color: var(--accent);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    margin-bottom: 4px;
}}

.header__sub {{
    font-size: 11px;
    color: var(--fg-dim);
}}

.stats {{
    display: flex;
    gap: 0;
    margin-bottom: 28px;
    border-bottom: 1px solid var(--rule);
    padding-bottom: 16px;
}}

.stat {{
    flex: 1;
    padding: 0 16px;
    border-right: 1px solid var(--rule);
}}
.stat:first-child {{ padding-left: 0; }}
.stat:last-child {{ border-right: none; }}

.stat__val {{
    font-size: 20px;
    font-weight: 700;
    color: var(--fg-bright);
    line-height: 1.2;
}}

.stat__label {{
    font-size: 10px;
    color: var(--fg-dim);
    letter-spacing: 0.06em;
    text-transform: uppercase;
}}

.day-mark {{
    font-size: 11px;
    color: var(--accent-dim);
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 16px 0 6px;
    border-bottom: 1px solid var(--rule);
    margin-bottom: 2px;
}}

.ev {{
    display: grid;
    grid-template-columns: 64px 1fr;
    gap: 12px;
    padding: 6px 0;
    border-bottom: 1px solid var(--rule);
    align-items: start;
}}

.ev__time {{
    font-size: 11px;
    color: var(--fg-dim);
    padding-top: 1px;
    text-align: right;
}}

.ev__body {{
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 6px;
}}

.ev__tag {{
    display: inline-block;
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.08em;
    padding: 1px 5px;
    border-radius: 2px;
    vertical-align: baseline;
}}

.ev__tag--register {{ background: var(--green); color: #0c1017; }}
.ev__tag--login {{ background: var(--amber); color: #0c1017; }}
.ev__tag--infra {{ background: var(--steel); color: #0c1017; }}
.ev__tag--restore {{ background: transparent; color: var(--steel); border: 1px solid var(--steel); }}
.ev__tag--stop {{ background: var(--red); color: #0c1017; }}
.ev__tag--issued {{ background: var(--violet); color: #0c1017; }}
.ev__tag--complete {{ background: var(--teal); color: #0c1017; }}

.ev__codename {{
    color: var(--green);
    font-weight: 600;
    font-size: 13px;
}}

.ev__cmd {{
    color: var(--accent);
    font-weight: 600;
}}

.ev__params {{
    color: var(--fg-bright);
}}

.ev__meta {{
    font-size: 12px;
    color: var(--fg);
}}

.ev__meta--dim {{
    color: var(--fg-dim);
    font-size: 11px;
}}

.ev__output-wrap {{
    width: 100%;
    margin-top: 4px;
}}

.ev__output-toggle {{
    font-size: 10px;
    color: var(--fg-dim);
    cursor: pointer;
    user-select: none;
    list-style: none;
}}

.ev__output-toggle::-webkit-details-marker {{ display: none; }}

.ev__output-toggle::before {{
    content: "\\25B8 ";
    color: var(--fg-dim);
}}

details[open] > .ev__output-toggle::before {{
    content: "\\25BE ";
}}

.ev__output {{
    background: var(--bg-output);
    color: var(--fg);
    font-size: 11px;
    line-height: 1.45;
    padding: 10px 12px;
    margin-top: 4px;
    border-left: 2px solid var(--rule);
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 280px;
    overflow-y: auto;
    overflow-x: auto;
}}

.ev--login {{
    opacity: 0.55;
}}

.ev--infra .ev__tag--restore {{
    opacity: 0.5;
}}
</style>

<div class="shell">
    <div class="header">
        <div class="header__title">Kronos &mdash; Event Timeline</div>
        <div class="header__sub">{datetime.now().strftime("%Y-%m-%d %H:%M:%S")} &middot; {len(events)} events</div>
    </div>

    <div class="stats">
        <div class="stat">
            <div class="stat__val">{n_agents}</div>
            <div class="stat__label">Agents</div>
        </div>
        <div class="stat">
            <div class="stat__val">{n_reg}</div>
            <div class="stat__label">Registrations</div>
        </div>
        <div class="stat">
            <div class="stat__val">{n_issued}</div>
            <div class="stat__label">Tasks Issued</div>
        </div>
        <div class="stat">
            <div class="stat__val">{n_complete}</div>
            <div class="stat__label">Completed</div>
        </div>
        <div class="stat">
            <div class="stat__val">{n_logins}</div>
            <div class="stat__label">Logins</div>
        </div>
    </div>

    <div class="timeline">
        {"".join(all_rows)}
    </div>
</div>
'''

with open(OUTPUT_FILE, "w") as f:
    f.write(html)

print(f"Timeline generated: {OUTPUT_FILE}")
print(f"Total events: {len(events)}")
