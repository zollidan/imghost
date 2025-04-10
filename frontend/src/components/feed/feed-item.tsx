import Image from "next/image"

type FeedItemProps = {
    src: string
}

export const FeedItem = ({ src }: FeedItemProps) => {
    return (
        <div className="h-fit bg-[#383838] rounded-xl overflow-hidden relative">
            <Image
                src={src}
                alt="Feed Item"
                width={300}
                height={400}
                className="object-cover"
            />
        </div>
    )
}