const Toggle = ({ fnClick, fnChange, title = "", checked = false }) => (
  <div className="flex items-center justify-center w-full mb-12">
    <label htmlFor="toogleA" className="flex items-center cursor-pointer">
      <div className="relative">
        <input
          id="toogleA"
          type="checkbox"
          className="sr-only"
          onClick={(e) => {
            if (fnClick !== undefined) fnClick(e.target.checked);
          }}
          onChange={(e) => {
            if (fnChange !== undefined) fnChange(e.target.checked);
          }}
          type="checkbox"
          checked={checked}
        />
        <div className="w-10 h-4 bg-gray-400 rounded-full shadow-inner"></div>
        <div className="dot absolute w-6 h-6 bg-white rounded-full shadow -left-1 -top-1 transition"></div>
      </div>
      <div className="ml-3 text-gray-700 font-medium">{title}</div>
    </label>
  </div>
);

export default Toggle;
