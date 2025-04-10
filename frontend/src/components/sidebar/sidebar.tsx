'use client'

import Link from "next/link"

import { Button } from "./button"

import { FiUser, FiBell, FiSettings, FiMessageSquare, FiHome } from "react-icons/fi"

export const Sidebar = () => {
    const cons = () => {
        console.log('print')
    }

    return (
        <aside className="absolute top-0 left-0 flex flex-col justify-between gap-6 p-3 inset-y-0 border-r border-[#383838] bg-[#202020] z-50">
            <div className="flex flex-col gap-6">
                <Link href='/' className="p-3 text-white hover:bg-[#383838] rounded-full transition-colors">
                    <FiHome className="size-6" />
                </Link>
                <Link href='/user' className="p-3 text-white hover:bg-[#383838] rounded-xl transition-colors">
                    <FiUser className="size-6" />
                </Link>
                <Button onClick={cons}>
                    <FiBell className="size-6" />
                </Button>
                <Button onClick={cons}>
                    <FiMessageSquare className="size-6" />
                </Button>
            </div>
            <Button onClick={cons}>
                <FiSettings className="size-6" />
            </Button>
        </aside>
    )
}