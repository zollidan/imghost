import type { Metadata } from "next"
import { Jost } from "next/font/google"
import { Searchbar } from "@/components/searchbar"
import { Sidebar } from "@/components/sidebar/sidebar"
import "./globals.css"

const jost = Jost({
    variable: "--font-jost",
    subsets: ["latin", "cyrillic"],
})

export const metadata: Metadata = {
    title: "Create Next App",
    description: "Generated by create next app",
}

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en">
            <body
                className={`${jost.className} antialiased`}
            >
                <main className="relative min-h-screen bg-[#202020]">
                    <Sidebar />
                    <div className="space-y-6 ml-24 py-3 pr-3">
                        <Searchbar />
                        {children}
                    </div>
                </main>
            </body>
        </html>
    )
}
