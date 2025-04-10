import { FiSearch } from "react-icons/fi"

export const Searchbar = () => {
    return (
        <div className="flex items-center justify-between w-full p-3 bg-[#383838] rounded-xl placeholder:text-red-500">
            <button className="rounded-xl">
                <FiSearch className="size-6 text-[#7E7E7E]" />
            </button>
            <input type="text" placeholder="Поиск" className="w-full px-2 text-white bg-transparent border-none outline-none" />
        </div>
    )
}