import React from "react";
import Table, {
  NameCell,
  StarCell,
  LanguageCell,
  GrowthCell,
  SelectColumnFilter,
  SliderColumnFilter,
  filterGreaterThan,
} from "./Table";
import { dataCompare, sumNumberObjects, massageData } from "./shared/Utils";
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
      filter: "equals",
    },
    {
      Header: "Stars",
      accessor: "Stars",
      Cell: StarCell,
      Filter: SliderColumnFilter,
      filter: filterGreaterThan,
      sortType: "number",
      sortDescFirst: true,
      disableGroupBy: true,
      aggregate: "sum",
      Aggregated: ({ value }) => `${value} (total)`,
    },
    {
      Header: (
        <span>
          <GraphIcon />
          &nbsp;30d
        </span>
      ),
      id: "Growth30",
      accessor: (stuff) => ({
        baseline: stuff.data.Baseline30,
        added: stuff.data.Added30,
      }),
      Cell: GrowthCell,
      disableGroupBy: true,
      Filter: false,
      sortType: dataCompare,
      sortDescFirst: true,
      aggregate: sumNumberObjects,
    },
    {
      Header: (
        <span>
          <GraphIcon />
          &nbsp;180d
        </span>
      ),
      id: "Growth180",
      accessor: (stuff) => ({
        baseline: stuff.data.Baseline180,
        added: stuff.data.Added180,
      }),
      Cell: GrowthCell,
      disableGroupBy: true,
      Filter: false,
      sortType: dataCompare,
      sortDescFirst: true,
      aggregate: sumNumberObjects,
    },
    {
      Header: (
        <span>
          <GraphIcon />
          &nbsp;365d
        </span>
      ),
      id: "Growth365",
      accessor: (stuff) => ({
        baseline: stuff.data.Baseline365,
        added: stuff.data.Added365,
      }),
      Cell: GrowthCell,
      disableGroupBy: true,
      Filter: false,
      sortType: dataCompare,
      sortDescFirst: true,
      aggregate: sumNumberObjects,
    },
  ];

  return (
    <div className="min-h-screen bg-gray-100 text-gray-900">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 pt-4">
        <div className="">
          <img className="h-24" src={logoSrc} />
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
