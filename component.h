#ifndef COMPONENT_H
#define COMPONENT_H

#include <Windows.h>

typedef const char *(*player_getNamePtr)(void *player);
typedef void (*player_sendClientMessagePtr)(void *player, int colour, const char *message);
typedef void *(*vehicle_createPtr)(int isStatic, int modelID, float x, float y, float z, float angle, int colour1, int colour2, int respawnDelay, int addSiren);

const char *player_getName(void *player);
void player_sendClientMessage(void *player, int colour, const char *message);
void *vehicle_create(int isStatic, int modelID, float x, float y, float z, float angle, int colour1, int colour2, int respawnDelay, int addSiren);

// 

void *loadLib(const char *name);
void unloadLib(void *handle);
void *findFunc(void *handle, const char *name);
void initFuncs(void *handle);

#endif // COMPONENT_H