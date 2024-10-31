import { Injectable } from '@angular/core';
import { Compress } from '../../wailsjs/go/services/CompressService';
import { interactions } from '../../wailsjs/go/models';

@Injectable({
  providedIn: 'root'
})
export class CompressService {
  async compress(fileInfo: interactions.GameDataInfo) {
    Compress(fileInfo);
  }
}
