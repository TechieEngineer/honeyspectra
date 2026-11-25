from fastapi.testclient import TestClient
from backend.app import app, Store
from simulator import generate
import random

def test_demo_enrichment():
    s=Store()
    for _ in range(20):
        for e in generate(random.Random(_)):
            s.add(e)
    assert s.summary()['total']>20
    assert 'T1110' in s.summary()['mitre']
    assert s.summary()['hashes']
    assert all(e['source_ip'].startswith('192.0.2.') for e in s.events)

def test_routes():
    with TestClient(app) as c:
        assert c.get('/health').json()['status']=='ok'
        assert c.get('/').status_code==200
        assert c.get('/api/events').status_code==200
        assert c.get('/api/summary').status_code==200
        assert c.get('/api/session/missing').json()==[]

def test_live_no_mitre_inference():
    s=Store();s.add({'protocol':'http','timestamp':'2026-01-01T00:00:00Z','session_id':'a','kind':'exploit_demo','synthetic':False});assert s.events[0]['mitre']==''
