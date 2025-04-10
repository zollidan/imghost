'use client'

import { Masonry } from "./masonry"
import { FeedItem } from "./feed-item"

const pics = ['https://img.freepik.com/free-photo/creative-light-bulb-abstract-glowing-blue-background-generative-ai_188544-8090.jpg?semt=ais_hybrid&w=740', 'https://img.freepik.com/free-photo/colombian-national-soccer-team-concept_23-2150257160.jpg?ga=GA1.1.139798330.1744297803&semt=ais_hybrid&w=740', 'https://img.freepik.com/free-photo/creative-composition-with-fruits-texture-vibrant-colors_23-2149888008.jpg?t=st=1744298796~exp=1744302396~hmac=9adcf9171e85d1377e6ba52ce04d3647ae19eb35b95b3834a3bb4270319e6956&w=740', 'https://img.freepik.com/free-photo/female-football-player-kicking-ball_23-2148850752.jpg?ga=GA1.1.139798330.1744297803&semt=ais_hybrid&w=740', 'https://img.freepik.com/free-photo/front-view-new-football-pedestal_23-2148796901.jpg?ga=GA1.1.139798330.1744297803&semt=ais_hybrid&w=740', 'https://img.freepik.com/free-photo/emotional-young-sportsman-make-sport-exercises-looking-aside_171337-15401.jpg?ga=GA1.1.139798330.1744297803&semt=ais_hybrid&w=740', 'https://img.freepik.com/free-photo/robot-hand-holding-lemon-bulb_53876-88558.jpg?t=st=1744298694~exp=1744302294~hmac=22ff94ae45c2886fed3a7d1dcc924de7e4b8279f5ec1cf626d6ab3a382083813&w=740', 'https://img.freepik.com/free-photo/beautiful-portrait-teenager-woman_23-2149453366.jpg?t=st=1744298721~exp=1744302321~hmac=0fedc5f9fc9b9158b61c089fd888f6796588ce9acec51e559a3d245a4f4a2d83&w=826', 'https://img.freepik.com/free-vector/creative-social-media-blogger-greek-statue-media-mix-post_53876-116532.jpg?t=st=1744298734~exp=1744302334~hmac=5b358c934949389dbfabbe6df125ae9b6ad752d4b7d06044f94dadf18b8acf9e&w=740', 'https://img.freepik.com/free-photo/anaglyph-effect-hand-holding-paper-plane_53876-126885.jpg?ga=GA1.1.139798330.1744297803&semt=ais_hybrid&w=740',]

export const Feed = () => {
    return (
        <Masonry
            breakpointCols={{ default: 6, 1024: 3, 768: 2, 480: 1 }}
            className="masonry-grid"
            columnClassName="masonry-column"
        >
            {pics.map((pic, index) => (
                <FeedItem key={index} src={pic} />
            ))}
        </Masonry>
    )
}