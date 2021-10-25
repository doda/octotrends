import React from "react";
import Table, {
  NameCell,
  StarCell,
  LanguageCell,
  GrowthCell,
  SelectColumnFilter,
  filterInRange,
  SizeFilter,
} from "./Table";
import {
  dataCompare,
  sumNumberObjects,
  massageData,
  equalsForSelect,
} from "./shared/Utils";
import { GraphIcon } from "@primer/octicons-react";
import JSONData from "./out.json";
import logoSrc from "./images/octotrends-logo-black.png";

function App() {
  const columns = [
    {
      Header: "Name",
      accessor: "Name",
      Cell: NameCell,
      disableGroupBy: true,
      Filter: false,
      display: false,
    },
    {
      Header: "Language",
      accessor: "Language",
      Filter: SelectColumnFilter,
      Cell: LanguageCell,
      filter: equalsForSelect,
    },
    {
      Header: "Stars",
      accessor: "Stars",
      Cell: StarCell,
      Filter: SizeFilter,
      filter: filterInRange,
      sortType: "number",
      sortDescFirst: true,
      disableGroupBy: true,
      aggregate: "sum",
      Aggregated: ({ value }) => `${value} (total)`,
    },
    ...
      [30, 180, 365].map((period) => ({
        Header: (
          <span
            title="Stars added over the last 180 days"
            className="whitespace-nowrap"
          >
            <GraphIcon /> {period}d
          </span>
        ),
        id: `Growth${period}`,
        accessor: (stuff) => ({
          baseline: stuff.data[`Baseline${period}`],
          added: stuff.data[`Added${period}`],
        }),
        Cell: GrowthCell,
        disableGroupBy: true,
        Filter: false,
        sortType: dataCompare,
        sortDescFirst: true,
        aggregate: sumNumberObjects,
      })),
    ,
  ];

  return (
    <div className="min-h-screen bg-gray-100 text-gray-900">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 pt-4">
        <div className="">
          <img className="h-24" alt="OctoTrends logo" src={logoSrc} />
          <h1 className="text-xl font-semibold">
            Find trending repositories on GitHub
          </h1>
        </div>
        <div className="mt-6">
          <Table columns={columns} data={massageData(JSONData)} />
        </div>
      </main>
    </div>
  );
}

export default App;
