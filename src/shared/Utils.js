export function classNames(...classes) {
  return classes.filter(Boolean).join(" ");
}

export function growthCalc(data) {
  if (data.baseline === 0) return 0;
  return (data.baseline + data.added) / data.baseline;
}

export function humanNumber(number) {
  if (number <= 1000) return `${number}`;
  let result, unit;
  if (number > 1000000) {
    result = `${(number / 1000000).toFixed(1)}`;
    unit = "M";
  } else if (number > 1000) {
    result = `${(number / 1000).toFixed(0)}`;
    unit = "k";
  }
  return parseFloat(result) + unit;
}

export function compareBasic(a, b) {
  return a === b ? 0 : a > b ? 1 : -1;
}

// export function dataCompare(rowA, rowB, columnId) {
//   let dataA = growthCalc(rowA.values[columnId]);
//   let dataB = growthCalc(rowB.values[columnId]);
//   return compareBasic(dataA, dataB);
// }

export function dataCompare(rowA, rowB, columnId) {
  let dataA = (rowA.values[columnId]).added;
  let dataB = (rowB.values[columnId]).added;
  return compareBasic(dataA, dataB);
}

export function sumNumberObjects(leafValues) {
  return leafValues.reduce((prev, cur) => {
    let newo = {};
    for (let key in prev) newo[key] = prev[key] + cur[key];
    return newo;
  });
}

export function massageData(data) {
  // Unfortunately have to combine added / baseline data into 1 value in
  // order to be able to groupby and recalculate growth rates
  return data.map((obj) => {
    return {
      Name: obj.Name,
      Stars: obj.Stars,
      Language: obj.Language,
      Description: obj.Description,
      data: {
        Added30: obj.Added30,
        Baseline30: obj.Baseline30,
        Added180: obj.Added180,
        Baseline180: obj.Baseline180,
        Added365: obj.Added365,
        Baseline365: obj.Baseline365,
      },
    };
  });
}


export function equalsForSelect(rows, ids, filterValue) {
  return rows.filter((row) => {
    return ids.some((id) => {
      const rowValue = row.values[id];
      // eslint-disable-next-line eqeqeq
      return rowValue !== "" && (rowValue == filterValue || filterValue === "");
    });
  });
};

equalsForSelect.autoRemove = (val) => val == null;
