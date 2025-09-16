import type { Analytics } from "../entities/analytics";
interface Props {
  data: Analytics[];
}

export const AnalyticsTable = ({ data }: Props) => (
  <table className="table-auto border-collapse border border-gray-300 w-full mt-4">
    <thead>
      <tr className="bg-gray-100">
        <th className="border px-2 py-1">Time</th>
        <th className="border px-2 py-1">Device</th>
        <th className="border px-2 py-1">OS</th>
        <th className="border px-2 py-1">Browser</th>
        <th className="border px-2 py-1">IP</th>
      </tr>
    </thead>
    <tbody>
      {data.map((e) => (
        <tr key={e.id}>
          <td className="border px-2 py-1">
            {new Date(e.createdAt).toLocaleString()}
          </td>
          <td className="border px-2 py-1">{e.device}</td>
          <td className="border px-2 py-1">{e.os}</td>
          <td className="border px-2 py-1">{e.browser}</td>
          <td className="border px-2 py-1">{e.ip}</td>
        </tr>
      ))}
    </tbody>
  </table>
);
