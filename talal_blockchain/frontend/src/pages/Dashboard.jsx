import React, { useEffect, useMemo, useState } from "react";
import { addTx, getChain, getPending, mine, searchTx } from "../api/blockchain";
import { Pickaxe, Plus, RefreshCw, Search } from "lucide-react";
import BlockCard from "../components/BlockCard";
import SectionHeader from "../components/SectionHeader";

export default function Dashboard() {
  const [loading, setLoading] = useState(false);
  const [chain, setChain] = useState(null);
  const [pending, setPending] = useState([]);
  const [tx, setTx] = useState("");
  const [q, setQ] = useState("");
  const [searchRes, setSearchRes] = useState(null);
  const [toast, setToast] = useState("");

  const chainBlocks = useMemo(() => chain?.chain ?? [], [chain]);

  async function refreshAll() {
    setLoading(true);
    setToast("");
    try {
      const [c, p] = await Promise.all([getChain(), getPending()]);
      setChain(c);
      setPending(p.pendingTxs || []);
    } catch (e) {
      setToast(e?.message || "Failed to load");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    refreshAll();
  }, []);

  async function onAddTx() {
    setLoading(true);
    setToast("");
    try {
      await addTx(tx);
      setTx("");
      await refreshAll();
      setToast("✅ Transaction added to pending pool");
    } catch (e) {
      setToast("❌ " + (e?.response?.data?.message || e?.message || "Failed to add tx"));
    } finally {
      setLoading(false);
    }
  }

  async function onMine() {
    setLoading(true);
    setToast("");
    try {
      await mine();
      await refreshAll();
      setToast("⛏️ Block mined and added to chain");
    } catch (e) {
      setToast("❌ " + (e?.response?.data?.message || e?.message || "Mining failed"));
    } finally {
      setLoading(false);
    }
  }

  async function onSearch() {
    if (!q.trim()) return;
    setLoading(true);
    setToast("");
    try {
      const res = await searchTx(q);
      setSearchRes(res);
    } catch (e) {
      setToast("❌ " + (e?.response?.data?.message || e?.message || "Search failed"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="container">
      <div className="card cardGlow" style={{ marginBottom: 14 }}>
        <div className="row" style={{ justifyContent: "space-between" }}>
          <div className="title">
            {chain?.name || "Talal Nadeem"} <span className="badge">Blockchain</span>
          </div>
          <div className="row">
            <button className="btn" onClick={refreshAll} disabled={loading}>
              <RefreshCw size={16} /> Refresh
            </button>
            <span className="badge">Difficulty: {chain?.difficulty ?? 3}</span>
          </div>
        </div>

        <div style={{ height: 12 }} />
        <div className="grid grid2">
          {/* Add TX */}
          <div className="card" style={{ borderRadius: 16 }}>
            <SectionHeader
              title="Add Transaction"
              right={<span className="badge">{pending.length} pending</span>}
            />
            <div style={{ height: 10 }} />
            <input
              className="input"
              placeholder='e.g. "Alice -> Bob : 5"'
              value={tx}
              onChange={(e) => setTx(e.target.value)}
            />
            <div style={{ height: 10 }} />
            <div className="row">
              <button className="btn btnPrimary" onClick={onAddTx} disabled={loading}>
                <Plus size={16} /> Add to Pool
              </button>
              <button className="btn" onClick={() => setTx("Alice -> Bob : 5")} disabled={loading}>
                Autofill
              </button>
            </div>

            <div style={{ height: 10 }} />
            <div className="muted">Pending Pool</div>
            <div className="toast" style={{ marginTop: 8 }}>
              {pending.length === 0 ? (
                <span className="muted">No pending transactions. Add one to mine.</span>
              ) : (
                <ul style={{ margin: 0, paddingLeft: 18 }}>
                  {pending.map((p, i) => (
                    <li key={i} className="mono" style={{ marginBottom: 6 }}>
                      {p}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>

          {/* Mine + Search */}
          <div className="card" style={{ borderRadius: 16 }}>
            <SectionHeader title="Mine + Search" />
            <div style={{ height: 10 }} />

            <button className="btn btnPrimary" onClick={onMine} disabled={loading}>
              <Pickaxe size={16} /> Mine Block
            </button>
            <div className="muted" style={{ marginTop: 8 }}>
              Mines the next block from pending transactions using Proof-of-Work.
            </div>

            <div style={{ height: 16 }} />

            <div className="muted">Search transactions</div>
            <div className="row" style={{ marginTop: 8 }}>
              <input
                className="input"
                placeholder="Search e.g. Bob"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                onKeyDown={(e) => (e.key === "Enter" ? onSearch() : null)}
              />
              <button className="btn" onClick={onSearch} disabled={loading}>
                <Search size={16} /> Search
              </button>
            </div>

            {searchRes && (
              <div className="toast" style={{ marginTop: 10 }}>
                <div className="muted">
                  Results for <span className="mono">{searchRes.query}</span> —{" "}
                  <span className="good">{searchRes.count}</span> match(es)
                </div>
                <div style={{ height: 8 }} />
                {searchRes.count === 0 ? (
                  <div className="muted">No matches found.</div>
                ) : (
                  <ul style={{ margin: 0, paddingLeft: 18 }}>
                    {searchRes.results.map((r, i) => (
                      <li key={i} className="mono" style={{ marginBottom: 6 }}>
                        [Block #{r.blockIndex}] {r.tx}
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}
          </div>
        </div>

        {toast && <div className="toast">{toast}</div>}
      </div>

      <div className="card cardGlow">
        <SectionHeader
          title="Blockchain"
          right={<span className="badge">{chainBlocks.length} blocks</span>}
        />
        <div style={{ height: 12 }} />
        <div className="grid">
          {chainBlocks.map((b) => (
            <BlockCard key={b.hash} block={b} />
          ))}
        </div>
      </div>
    </div>
  );
}