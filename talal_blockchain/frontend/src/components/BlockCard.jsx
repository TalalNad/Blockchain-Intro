import React from "react";

export default function BlockCard({ block }) {
  return (
    <div className="card cardGlow" style={{ animation: "floaty 6s ease-in-out infinite" }}>
      <div className="row" style={{ justifyContent: "space-between" }}>
        <div className="title">
          Block <span className="badge">#{block.index}</span>
        </div>
        <span className="badge">Nonce: {block.nonce}</span>
      </div>

      <div className="muted" style={{ marginTop: 8 }}>
        Timestamp: <span className="mono">{block.timestamp}</span>
      </div>

      <div style={{ height: 12 }} />

      <div className="kv">
        <div>PrevHash</div>
        <div className="mono">{block.prevHash}</div>

        <div>MerkleRoot</div>
        <div className="mono">{block.merkleRoot}</div>

        <div>Hash</div>
        <div className="mono">{block.hash}</div>
      </div>

      <div style={{ height: 12 }} />
      <div className="muted">Transactions</div>
      <ul style={{ margin: "8px 0 0", paddingLeft: 18 }}>
        {block.txs?.map((tx, i) => (
          <li key={i} className="mono" style={{ marginBottom: 6 }}>
            {tx}
          </li>
        ))}
      </ul>
    </div>
  );
}