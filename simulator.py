"""Generate fictional event records only. No network traffic or external calls."""
import argparse
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import random
import time
import uuid

SCENARIOS = [
    ('ssh', [('auth_demo', 'Fictional login attempt'), ('shell_demo', '$ whoami → demo-user'), ('shell_demo', '$ pwd → /demo')]),
    ('http', [('exploit_demo', 'Fictional admin-route probe'), ('payload_demo', 'Harmless fixture hash')]),
    ('telnet', [('auth_demo', 'Fictional IoT login'), ('shell_demo', '$ ls → example.txt')]),
    ('ftp', [('auth_demo', 'Fictional FTP login'), ('payload_demo', 'Harmless fixture hash')]),
    ('mysql', [('auth_demo', 'Fictional database login')]),
    ('tcp', [('probe', 'Fictional service probe')]),
]

def generate(rng):
    protocol, steps = rng.choice(SCENARIOS)
    session = uuid.uuid4().hex
    ip = f'192.0.2.{rng.randint(1, 254)}'
    country, asn = rng.choice([('India', 'AS64512'), ('Germany', 'AS64513'), ('Japan', 'AS64514'), ('Brazil', 'AS64515'), ('United States', 'AS64516')])
    for kind, detail in steps:
        event = dict(session_id=session, timestamp=datetime.now(timezone.utc).isoformat(), source_ip=ip, protocol=protocol, kind=kind, detail=detail, synthetic=True, country=country, asn=asn)
        if kind == 'auth_demo':
            event.update(username=rng.choice(['demo-admin', 'demo-user', 'demo-guest']), password=rng.choice(['fixture-A', 'fixture-B', 'fixture-C']))
        if kind == 'payload_demo':
            event['payload_hash'] = hashlib.sha256(b'honeypot-core harmless demonstration fixture').hexdigest()
        yield event

def main():
    p = argparse.ArgumentParser(); p.add_argument('--count', type=int, default=100); p.add_argument('--interval', type=float, default=.25); p.add_argument('--output', type=Path, default=Path('data/demo.jsonl')); p.add_argument('--seed', type=int, default=42); args = p.parse_args()
    if not 1 <= args.count <= 50000 or not 0 <= args.interval <= 60:
        p.error('count 1..50000, interval 0..60')
    args.output.parent.mkdir(parents=True, exist_ok=True)
    rng = random.Random(args.seed)
    with args.output.open('a', encoding='utf-8') as f:
        for _ in range(args.count):
            for event in generate(rng):
                f.write(json.dumps(event)+'\n')
            f.flush()
            time.sleep(args.interval)
    print(f'Wrote {args.count} fictional sessions to {args.output}')
if __name__ == '__main__': main()
