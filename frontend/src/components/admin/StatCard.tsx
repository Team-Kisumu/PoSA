interface Props {
  label: string;
  value: number | string;
  icon: React.ReactNode;
  color: "emerald" | "blue" | "purple" | "amber" | "red";
  suffix?: string;
  trend?: { value: string; up: boolean };
}

const colorMap = {
  emerald: { bg: "bg-emerald-500/10", text: "text-emerald-400", ring: "ring-emerald-500/20" },
  blue: { bg: "bg-blue-500/10", text: "text-blue-400", ring: "ring-blue-500/20" },
  purple: { bg: "bg-purple-500/10", text: "text-purple-400", ring: "ring-purple-500/20" },
  amber: { bg: "bg-amber-500/10", text: "text-amber-400", ring: "ring-amber-500/20" },
  red: { bg: "bg-red-500/10", text: "text-red-400", ring: "ring-red-500/20" },
};

export default function StatCard({ label, value, icon, color, suffix = "", trend }: Props) {
  const c = colorMap[color];
  return (
    <div className="bg-[#0c1222] border border-white/[0.06] rounded-2xl p-5 hover:border-white/[0.1] transition-colors">
      <div className="flex items-start justify-between mb-4">
        <div className={`w-11 h-11 rounded-xl flex items-center justify-center ${c.bg} ring-1 ${c.ring}`}>
          <span className={c.text}>{icon}</span>
        </div>
        {trend && (
          <span className={`text-[11px] font-medium px-2 py-0.5 rounded-full ${trend.up ? "bg-emerald-500/10 text-emerald-400" : "bg-red-500/10 text-red-400"}`}>
            {trend.up ? "\u2191" : "\u2193"} {trend.value}
          </span>
        )}
      </div>
      <p className="text-[28px] font-bold text-white leading-none">{value}{suffix}</p>
      <p className="text-[12px] text-zinc-500 mt-1.5">{label}</p>
    </div>
  );
}
