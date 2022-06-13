# terrain-generation

This program is a progressive perlin noise terrain generator written in golang using g3n: A golang bindings library for using OpenGL (https://github.com/g3n/engine)
It is currently configured for laying two terrains on top of each other, these terrains are called Bipartite Terrains. However, there is skeleton code for simpler terrains should they be desired
Keep an eye out for improvements! Especially in progressiveley generating terrain as the x and y displacements are modified

## How to use 

 - Make sure that your system can run g3n, dependancies and instructions can be found here: https://github.com/g3n/engine
 - Navigate to /Terrain-Generation and execute:
 $ ./run.sh
 - wait for the GUI to generate the terrain, you can navigate the terrain by scrolling the x and y meters. 
 - The GUI is set up with standard orbital controls for OpenGL: so you can use mouse scroll to Zoom, right-click to probe about the terrain, and left click to slide the camera

## Changing the Terrain

 - The seed for the terrain generator can be changed by modifying the SEED_# constants in main.go, It is suggested that prime numbers are used
 - The granularity of the rendered terrain can be changed by modifying the TERRAIN_WIDTH and TERRAIN_HEIGHT constants in main.go. Increasing the width or height will increase the number of triangles renders whereas decreasing the width or height will do the opposite
 - The number of gradients used by the Perlin Noise algorithm can be modified by changing the GRADIENT_WIDHT_B# and GRADIENT_HEIGHT_B# constants in main.go. These constants are used to determine how many gradients will exist within the rendered terrain, this will change the number of peaks and valleys seen in the terrain
 - To change the magnitude or the amplitude of the terrain generated, the M constant can be modified in main.go
 - TO change the significance of a Bipartite Terrains macro and micro noises, the PROPORITON constant can be modified in main.go
