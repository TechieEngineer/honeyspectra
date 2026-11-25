import asyncio
from collections import Counter, deque
from contextlib import asynccontextmanager
from datetime import datetime, timezone
import json
import os
from pathlib import Path
from fastapi import FastAPI, Request
from fastapi.responses import FileResponse, StreamingResponse
from fastapi.staticfiles import StaticFiles

DATA = Path(os.getenv('HONEYPOT_DATA', 'data'))
STATIC = Path(__file__).parent / 'static'
LIMIT = 50_000
TAGS = {'auth_demo': 'T1110', 'shell_demo': 'T1059', 'exploit_demo': 'T1190'}
PROTOCOLS = {'ssh', 'http', 'telnet', 'ftp', 'mysql', 'tcp'}

class Store:
    def __init__(self):
        self.events = deque(maxlen=LIMIT)
        self.seq = 0
    def add(self, event):
        if not isinstance(event, dict) or event.get('protocol') not in PROTOCOLS or not isinstance(event.get('session_id'), str):
            return
        if not isinstance(event.get('timestamp'), str):
            return
        self.seq += 1
        event = dict(event)
        event['seq'] = self.seq
        event['mitre'] = TAGS.get(event.get('kind'), '') if event.get('synthetic') is True else ''
        self.events.append(event)
    def summary(self):
        events = list(self.events)
        now = datetime.now(timezone.utc)
        recent = []
        for e in reversed(events):
            try:
                age = (now - datetime.fromisoformat(e['timestamp'].replace('Z', '+00:00'))).total_seconds()
                if 0 <= age <= 60:
                    recent.append(e)
            except (ValueError, TypeError):
                continue
        def count(key, rows=events):
            return dict(Counter(e[key] for e in rows if isinstance(e.get(key), str) and e[key]).most_common(12))
        demo = [e for e in events if e.get('synthetic') is True]
        return {'total': len(events), 'rate': round(len(recent)/60, 2), 'recent_sessions': len({e['session_id'] for e in recent}), 'unique_ips': len({e.get('source_ip') for e in events if isinstance(e.get('source_ip'), str)}), 'protocols': count('protocol'), 'mitre': count('mitre'), 'countries': count('country', demo), 'asns': count('asn', demo), 'usernames': count('username', demo), 'passwords': count('password', demo), 'hashes': count('payload_hash', demo)}

async def tail(store, directory):
    offsets = {}
    while True:
        for path in (directory/'engine.jsonl', directory/'demo.jsonl'):
            try:
                stat = path.stat()
                identity = (stat.st_dev, stat.st_ino)
                old_id, offset = offsets.get(path, (identity, max(0, stat.st_size - 16*1024*1024)))
                if old_id != identity or offset > stat.st_size:
                    offset = 0
                with path.open('rb') as f:
                    f.seek(offset)
                    if path not in offsets and offset:
                        f.readline(65536)
                    for _ in range(1000):
                        start = f.tell()
                        line = f.readline(65537)
                        if not line:
                            break
                        if len(line) > 65536:
                            # Discard the oversized chunk; malformed fragments will be ignored.
                            continue
                        if not line.endswith(b'\n'):
                            f.seek(start)
                            break
                        try:
                            store.add(json.loads(line))
                        except (ValueError, UnicodeError, TypeError):
                            pass
                    offsets[path] = (identity, f.tell())
            except OSError:
                pass
        await asyncio.sleep(.25)

@asynccontextmanager
async def lifespan(app):
    app.state.store = Store()
    task = asyncio.create_task(tail(app.state.store, DATA))
    try:
        yield
    finally:
        task.cancel()
        await asyncio.gather(task, return_exceptions=True)

app = FastAPI(title='Honeypot Core', lifespan=lifespan)
app.mount('/static', StaticFiles(directory=STATIC), name='static')

@app.get('/')
def home():
    return FileResponse(STATIC/'index.html')

@app.get('/health')
def health():
    return {'status': 'ok', 'mode': 'local-lab'}

@app.get('/api/summary')
async def summary(request: Request):
    return request.app.state.store.summary()

@app.get('/api/events')
async def events(request: Request, protocol: str = '', offset: int = 0, limit: int = 100):
    rows = [e for e in reversed(request.app.state.store.events) if not protocol or e['protocol'] == protocol]
    offset, limit = max(0, offset), min(200, max(1, limit))
    return {'total': len(rows), 'events': rows[offset:offset+limit]}

@app.get('/api/session/{session_id}')
async def session(session_id: str, request: Request):
    return [e for e in request.app.state.store.events if e['session_id'] == session_id][:200]

@app.get('/api/stream')
async def stream(request: Request):
    async def generate():
        cursor = request.app.state.store.seq
        yield 'data: ' + json.dumps({'seq': cursor}) + '\n\n'
        heartbeat = 0
        while not await request.is_disconnected():
            await asyncio.sleep(.35)
            seq = request.app.state.store.seq
            heartbeat += 1
            if seq != cursor or heartbeat >= 30:
                cursor, heartbeat = seq, 0
                yield 'data: ' + json.dumps({'seq': cursor}) + '\n\n'
    return StreamingResponse(generate(), media_type='text/event-stream', headers={'Cache-Control': 'no-cache', 'X-Accel-Buffering': 'no'})
