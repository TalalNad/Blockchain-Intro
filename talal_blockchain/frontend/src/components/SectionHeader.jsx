import React from "react";

export default function SectionHeader({ title, right }) {
  return (
    <div className="row" style={{ justifyContent: "space-between" }}>
      <div className="title">{title}</div>
      <div className="row">{right}</div>
    </div>
  );
}