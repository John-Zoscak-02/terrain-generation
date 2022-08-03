# terrain-generation

![alt text](https://github.com/John-Zoscak-02/terrain-generation/blob/main/image/terrain-generation.png?raw=true)

This program is a progressive perlin noise terrain generator written in golang using g3n: A golang bindings library for using OpenGL (https://github.com/g3n/engine)

## How to use 

 - Make sure that your system can run g3n, dependancies and instructions can be found here: https://github.com/g3n/engine
 - Necessary audio DLLs for windows are in /audiodlls 
 - Navigate to /Terrain-Generation and execute:
 $ ./run.sh <mapname> <terrain_width> <terrain_height>
    - mapname: The program will navigate to the /maps directory in the project and will search for <mapname>.json to render it
    - terrain_width: The number of triangles you want rendered in the x-direction
    - terrain_height: The number of triangles you want rendered inthe y-direction
 - wait for a GUI with the terrain to pop up, you can navigate the terrain by scrolling the x and y meters at the left of the GUI. 
 - The GUI is set up with standard orbital controls for OpenGL: so you can use mouse scroll to Zoom, right-click to probe about the terrain, and left click to slide the camera

## Changing the Terrain

 - There are two types of terrains that can be generated, simple terrains and bipartite terrains. You can specify the type of terrain that you want by changing the typ value in the map's json file
    - (typ=1) Simple terrains are a basic perlin noise terrain
    - (typ=2) Bipartite terrains use an independent macro and micro perlin noise generators to produce more unique textures
 - The seed for the terrain generator can be changed by modifying the seed_# in the json file for the map that you are generating, It is suggested that prime numbers are used
 - The granularity of the rendered terrain can be changed by modifying the terrain_width and terrain_height command line arguments.
 - The number of gradients used by the Perlin Noise algorithm in the area of rendered terrain can be modified by changing the gradient_width_b# and gradient_height_b# values in the json files of the map.
 - To change the magnitude or the amplitude of the terrain generated, the m value can be modified in the map's json
 - To change the significance of a Bipartite Terrain's macro and micro noises, the prop value can be modified in the map's json
